package app

import (
	"context"
	"time"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

func initTickers(ctx context.Context, services *service.Services, conf *config.Config, _ storage.IStorage, l logger.Logger) {
	if services.ExRateService != nil {
		go services.ExRateService.Run(ctx)
	}

	if services.WebHookService != nil {
		go services.WebHookService.Run(ctx)
	}

	if services.WithdrawService != nil {
		go services.WithdrawService.Run(ctx, models.AllBlockchain())
	}

	if services.UnconfirmedCollapser != nil {
		go services.UnconfirmedCollapser.Run(ctx, conf.Transactions.UnconfirmedCollapseInterval)
	}

	if services.NotificationService != nil {
		go services.NotificationService.Run(ctx)
	}

	go processingPingMonitor(ctx, services, l)

	if services.ExchangeService != nil {
		go services.ExchangeService.Run(ctx)
	}

	if services.ExchangeWithdrawalService != nil {
		go services.ExchangeWithdrawalService.RunWithdrawalQueue(ctx)
		go services.ExchangeWithdrawalService.RunWithdrawalUpdater(ctx)
	}

	if services.ExchangeRulesService != nil {
		go services.ExchangeRulesService.Run(ctx)
	}

	if services.SystemService != nil {
		go services.SystemService.RunHeartbeatLoop(ctx)
	}

	if services.BalanceUpdater != nil {
		go services.BalanceUpdater.Run(ctx, conf.Wallets.UpdateBalancesInterval)
	}

	if services.WalletBalanceService != nil {
		go services.WalletBalanceService.ProcessingBalanceStatsInBackground(ctx, conf.Wallets.UpdateTronResourcesInterval)
	}

	if services.AMLStatusChecker != nil {
		go services.AMLStatusChecker.Run(ctx)
	}
}

func processingPingMonitor(ctx context.Context, services *service.Services, l logger.Logger) {
	tickTocker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tickTocker.C:
			if services.ProcessingService.Initialized() {
				process, err := services.LogService.StartProcess(ctx, "PingProcessing")
				if err != nil {
					return
				}
				l.Debug("Ping Processing started", &logger.LogPrams{Status: logger.InProgress, ProcessID: process.ID, Slug: process.TypeSlug})
				if _, err := services.ProcessingSystemService.GetProcessingSystemInfo(ctx); err != nil {
					l.Debug("Processing err:", err, &logger.LogPrams{Status: logger.Failed, ProcessID: process.ID, Slug: process.TypeSlug})
				} else {
					l.Debug("Ping Processing finished", &logger.LogPrams{Status: logger.Completed, ProcessID: process.ID, Slug: process.TypeSlug})
				}
				_ = services.LogService.StopProcess(ctx, process.ID)
			}
		}
	}
}
