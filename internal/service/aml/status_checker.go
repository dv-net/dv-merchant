package aml

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_check_history"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_check_queue"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_checks"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_aml_settings"
	amlproviders "github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// StatusChecker polls enqueued aml_checks for status updates
type StatusChecker interface {
	Run(ctx context.Context)
}

var _ StatusChecker = (*Service)(nil)

const maxWorkers = 50

func (s *Service) Run(ctx context.Context) {
	if s.maxAttempts <= 0 {
		s.log.Warnw("aml status checker", "error", fmt.Errorf("max_attempts must be positive"))
	}

	s.checkInProgress.Store(false)

	go s.processQueue(ctx)

	ticker := time.NewTicker(s.checkStatusInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			go s.processQueue(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) processQueue(ctx context.Context) {
	if !s.checkInProgress.CompareAndSwap(false, true) {
		return
	}
	defer s.checkInProgress.Store(false)

	queue, err := s.st.AmlCheckQueue().FetchPending(ctx, s.maxAttempts, models.AmlCheckStatusPending)
	if err != nil {
		s.log.Errorw("failed to fetch pending checks", "error", err)
		return
	}

	wg := sync.WaitGroup{}
	sema := make(chan struct{}, maxWorkers)
	for _, check := range queue {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case sema <- struct{}{}:
		}

		wg.Add(1)
		go func(check *repo_aml_check_queue.FetchPendingRow) {
			defer wg.Done()
			defer func() { <-sema }()
			defer func() {
				if r := recover(); r != nil {
					s.log.Errorw("panic while processing aml check", "recover", r, "check_id", check.AmlCheck.ID)
				}
			}()

			if procErr := s.processCheckQueueElement(ctx, check); procErr != nil {
				s.log.Errorw("failed to process check", "error", procErr, "check_id", check.AmlCheck.ID)
			}
		}(check)
	}

	wg.Wait()
}

func (s *Service) processCheckQueueElement(ctx context.Context, check *repo_aml_check_queue.FetchPendingRow) error {
	providerSlug, ok := slugMapping[check.AmlService.Slug]
	if !ok {
		return fmt.Errorf("unsupported provider: %s", check.AmlService.Slug)
	}

	client, err := s.factory.GetClient(providerSlug)
	if err != nil {
		return fmt.Errorf("failed to get provider for %s: %w", check.AmlService.Slug, err)
	}

	var dto amlproviders.InitCheckDTO
	if err := json.Unmarshal(check.AmlCheckQueue.RequestPayload, &dto); err != nil {
		return fmt.Errorf("failed to unmarshal check payload: %w", err)
	}

	externalID := check.AmlCheck.ExternalID
	authExternalID := externalID
	if isPendingExternalID(externalID) {
		externalID = ""
		authExternalID = dto.TxID
	}

	_, authorizer, err := s.prepareServiceDataByUser(ctx, check.User.ID, prepareParams{Slug: check.AmlService.Slug, ExternalID: authExternalID})
	if err != nil {
		return fmt.Errorf("failed to prepare service '%s' data: %w", check.AmlService.Slug, err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.checkTimeout)
	defer cancel()

	externalCheckResult, err := client.Check(ctxWithTimeout, dto, externalID, authorizer)
	return s.handleCheckResult(ctx, check, externalCheckResult, err)
}

func (s *Service) handleCheckResult(
	ctx context.Context,
	check *repo_aml_check_queue.FetchPendingRow,
	result *amlproviders.CheckResponse,
	fetchErr error,
) error {
	var resolution checkResolution

	txErr := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Reset so a re-executed tx body can never fire a stale event.
		resolution = checkResolution{}

		if err := s.createCheckHistory(ctx, tx, check, result, fetchErr); err != nil {
			return fmt.Errorf("failed to create check history: %w", err)
		}

		if fetchErr != nil {
			if reqErr, ok := errors.AsType[*amlproviders.RequestFailedError](fetchErr); ok && !reqErr.Retryable {
				res, err := s.finalizeCheck(ctx, tx, check, models.AmlCheckStatusFailed, decimal.Zero, nil, nil)
				resolution = res
				return err
			}
			res, err := s.continueOrFailCheck(ctx, tx, check, decimal.Zero, nil)
			resolution = res
			return err
		}

		if result.ExternalID != "" && result.ExternalID != check.AmlCheck.ExternalID {
			if err := s.st.AmlChecks(repos.WithTx(tx)).UpdateExternalID(ctx, check.AmlCheck.ID, result.ExternalID); err != nil {
				return fmt.Errorf("failed to update external id: %w", err)
			}
			check.AmlCheck.ExternalID = result.ExternalID
		}

		resolvedStatus := convertAmlStatusToModel(result.Status)
		if resolvedStatus == models.AmlCheckStatusPending {
			res, err := s.continueOrFailCheck(ctx, tx, check, result.Score, result.Signals)
			resolution = res
			return err
		}

		riskLevel, err := convertAmlRiskLevelToModel(*result.RiskLevel)
		if err != nil {
			return fmt.Errorf("failed to convert risk level: %w", err)
		}

		res, err := s.finalizeCheck(ctx, tx, check, resolvedStatus, result.Score, riskLevel, result.Signals)
		resolution = res
		return err
	})
	if txErr != nil {
		return txErr
	}

	// The DB side of the verdict is already committed; mirror the dirty flag in
	// processing before telling anyone the check is done.
	s.SyncDirtyAddresses(ctx, check.User.ID, resolution.dirtyAddresses)

	if resolution.event != nil {
		s.dispatchCompletedEvent(*resolution.event)
	}

	return nil
}

func (s *Service) dispatchCompletedEvent(ev CheckCompletedEvent) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.log.Errorw("panic in aml check completed handler", "recover", r, "check_id", ev.Check.ID)
			}
		}()

		if err := s.eventListener.Fire(ev); err != nil {
			s.log.Errorw("aml check completed event handler failed",
				"error", err,
				"check_id", ev.Check.ID,
				"transaction_id", ev.Check.TransactionID.UUID,
				"status", ev.Check.Status,
			)
		}
	}()
}

// checkResolution carries what handleCheckResult must do once the transaction
// commits. A zero value means "nothing further to do".
type checkResolution struct {
	// event is set when the check reached a terminal state and its completion
	// event must be fired after the transaction commits.
	event *CheckCompletedEvent
	// dirtyAddresses were marked dirty in the DB and must be synced to processing
	// after the transaction commits.
	dirtyAddresses []*models.WalletAddress
}

func (s *Service) continueOrFailCheck(ctx context.Context, tx pgx.Tx, check *repo_aml_check_queue.FetchPendingRow, score decimal.Decimal, signals []amlproviders.SignalContribution) (checkResolution, error) {
	if check.IsLastAttempt {
		return s.finalizeCheck(ctx, tx, check, models.AmlCheckStatusFailed, score, nil, signals)
	}

	if err := s.st.AmlCheckQueue(repos.WithTx(tx)).IncrementAttempts(ctx, check.AmlCheckQueue.ID); err != nil {
		return checkResolution{}, fmt.Errorf("failed to increment attempts: %w", err)
	}

	return checkResolution{}, nil
}

func (s *Service) finalizeCheck(
	ctx context.Context,
	tx pgx.Tx,
	check *repo_aml_check_queue.FetchPendingRow,
	status models.AMLCheckStatus,
	score decimal.Decimal,
	riskLevel *models.AmlRiskLevel,
	signals []amlproviders.SignalContribution,
) (checkResolution, error) {
	if err := s.st.AmlChecks(repos.WithTx(tx)).UpdateAMLCheck(ctx, repo_aml_checks.UpdateAMLCheckParams{
		ID:        check.AmlCheck.ID,
		Status:    status,
		Score:     score,
		RiskLevel: riskLevel,
	}); err != nil {
		return checkResolution{}, fmt.Errorf("failed to update aml check to %s: %w", status, err)
	}

	if err := s.st.AmlCheckQueue(repos.WithTx(tx)).Delete(ctx, check.AmlCheckQueue.ID); err != nil {
		return checkResolution{}, fmt.Errorf("failed to delete from queue: %w", err)
	}

	s.log.Debugw("finalized check", "check_id", check.AmlCheck.ID, "status", status, "attempts", check.AmlCheckQueue.Attempts+1)

	if !check.AmlCheck.TransactionID.Valid {
		return checkResolution{}, nil
	}

	updatedCheck := check.AmlCheck
	updatedCheck.Status = status
	updatedCheck.Score = score
	updatedCheck.RiskLevel = riskLevel

	// Apply the verdict in the same transaction that removes the check from the queue.
	// While queued, the deposit address is excluded from transfers; if the verdict ran
	// after commit, a transfer could pick the address up in between.
	verdict, err := s.applyVerdictForCheck(ctx, tx, check, &updatedCheck, signals)
	if err != nil {
		return checkResolution{}, err
	}

	return checkResolution{
		event:          &CheckCompletedEvent{Check: updatedCheck, Signals: signals, Blocked: verdict.Blocked},
		dirtyAddresses: verdict.DirtyAddresses,
	}, nil
}

func (s *Service) applyVerdictForCheck(
	ctx context.Context,
	tx pgx.Tx,
	check *repo_aml_check_queue.FetchPendingRow,
	updatedCheck *models.AmlCheck,
	signals []amlproviders.SignalContribution,
) (VerdictResult, error) {
	trx, err := s.st.Transactions(repos.WithTx(tx)).GetById(ctx, check.AmlCheck.TransactionID.UUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Errorw("aml verdict skipped: transaction not found", "check_id", check.AmlCheck.ID, "tx_id", check.AmlCheck.TransactionID.UUID)
			return VerdictResult{}, nil
		}
		return VerdictResult{}, fmt.Errorf("fetch transaction for aml verdict: %w", err)
	}

	rules, err := s.st.UserAmlSettings(repos.WithTx(tx)).ListRiskRulesByUserID(ctx, repo_user_aml_settings.ListRiskRulesByUserIDParams{
		UserID:       trx.UserID,
		ProviderSlug: check.AmlService.Slug.String(),
	})
	if err != nil {
		return VerdictResult{}, fmt.Errorf("fetch aml risk rules: %w", err)
	}

	return s.ApplyVerdict(ctx, ApplyVerdictDTO{
		Check:         updatedCheck,
		UserID:        trx.UserID,
		StoreID:       trx.StoreID.UUID,
		TransactionID: trx.ID,
		WalletID:      trx.WalletID,
		ToAddress:     trx.ToAddress,
		Signals:       signals,
		Rules:         rules,
		DBTx:          tx,
	})
}

func (s *Service) createCheckHistory(
	ctx context.Context,
	tx pgx.Tx,
	check *repo_aml_check_queue.FetchPendingRow,
	result *amlproviders.CheckResponse,
	fetchErr error,
) error {
	params := repo_aml_check_history.CreateParams{
		AmlCheckID:    check.AmlCheck.ID,
		AttemptNumber: check.AmlCheckQueue.Attempts + 1,
	}

	params.RequestPayload = json.RawMessage(`{}`)
	params.ServiceResponse = json.RawMessage(`{}`)

	if fetchErr != nil {
		params.ErrorMsg = pgtypeutils.EncodeText(new(fetchErr.Error()))
	}
	if result != nil && result.Response != nil {
		params.ServiceResponse = result.Response
	}
	if result != nil && result.Request != nil {
		params.RequestPayload = result.Request
	}

	_, err := s.st.AmlCheckHistory(repos.WithTx(tx)).Create(ctx, params)
	return err
}
