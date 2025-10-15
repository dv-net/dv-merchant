package aml

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_check_history"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_check_queue"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_checks"
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

	go s.processQueue(ctx)

	ticker := time.NewTicker(s.checkStatusInterval)
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
	wg.Add(len(queue))

	sema := make(chan struct{}, maxWorkers)
	for _, check := range queue {
		sema <- struct{}{}
		go func() {
			defer wg.Done()
			if err = s.processCheckQueueElement(ctx, check); err != nil {
				s.log.Errorw("failed to process check", "error", err, "check_id", check.AmlCheck.ID)
			}

			<-sema
		}()
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

	_, authorizer, err := s.prepareServiceDataByUser(ctx, &check.User, prepareParams{Slug: check.AmlService.Slug, ExternalID: check.AmlCheck.ExternalID})
	if err != nil {
		return fmt.Errorf("failed to prepare service '%s' data: %w", check.AmlService.Slug, err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.checkTimeout)
	defer cancel()

	externalCheckResult, err := client.FetchCheckStatus(ctxWithTimeout, check.AmlCheck.ExternalID, authorizer)
	return s.handleCheckResult(ctx, check, externalCheckResult, err)
}

func (s *Service) handleCheckResult(
	ctx context.Context,
	check *repo_aml_check_queue.FetchPendingRow,
	result *amlproviders.CheckResponse,
	fetchErr error,
) error {
	return repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.createCheckHistory(ctx, tx, check, result, fetchErr); err != nil {
			return fmt.Errorf("failed to create check history: %w", err)
		}

		if fetchErr != nil {
			return s.continueOrFailCheck(ctx, tx, check, decimal.Zero)
		}

		resolvedStatus := convertAmlStatusToModel(result.Status)
		if resolvedStatus == models.AmlCheckStatusPending {
			return s.continueOrFailCheck(ctx, tx, check, result.Score)
		}

		riskLevel, err := convertAmlRiskLevelToModel(*result.RiskLevel)
		if err != nil {
			return fmt.Errorf("failed to convert risk level: %w", err)
		}

		return s.updateCheckAndClearQueue(ctx, tx, check, resolvedStatus, result.Score, riskLevel)
	})
}

// continueOrFailCheck increments attempts or finalizes check as failed if max_attempts was reached
func (s *Service) continueOrFailCheck(ctx context.Context, tx pgx.Tx, check *repo_aml_check_queue.FetchPendingRow, score decimal.Decimal) error {
	if check.IsLastAttempt {
		return s.updateCheckAndClearQueue(ctx, tx, check, models.AmlCheckStatusFailed, score, nil)
	}

	if err := s.st.AmlCheckQueue(repos.WithTx(tx)).IncrementAttempts(ctx, check.AmlCheckQueue.ID); err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	return nil
}

// updateCheckAndClearQueue finalize and complete queue element
func (s *Service) updateCheckAndClearQueue(
	ctx context.Context,
	tx pgx.Tx,
	check *repo_aml_check_queue.FetchPendingRow,
	status models.AMLCheckStatus,
	score decimal.Decimal,
	riskLevel *models.AmlRiskLevel,
) error {
	if err := s.st.AmlChecks(repos.WithTx(tx)).UpdateAMLCheck(ctx, repo_aml_checks.UpdateAMLCheckParams{
		ID:        check.AmlCheck.ID,
		Status:    status,
		Score:     score,
		RiskLevel: riskLevel,
	}); err != nil {
		return fmt.Errorf("failed to update aml check to %s: %w", status, err)
	}

	if err := s.st.AmlCheckQueue(repos.WithTx(tx)).Delete(ctx, check.AmlCheckQueue.ID); err != nil {
		return fmt.Errorf("failed to delete from queue: %w", err)
	}

	s.log.Debugw("finalized check", "check_id", check.AmlCheck.ID, "status", status, "attempts", check.AmlCheckQueue.Attempts+1)

	return nil
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
		errMsg := fetchErr.Error()
		params.ErrorMsg = pgtypeutils.EncodeText(&errMsg)
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
