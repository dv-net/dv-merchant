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

func buildWorkers(services *service.Services, conf *config.Config, _ storage.IStorage, l logger.Logger) []worker {
	var workers []worker
	add := func(name string, svcEnabled bool, run func(context.Context)) {
		if !svcEnabled {
			return
		}
		workers = append(workers, asWorker(name, run))
	}

	add("exrate", services.ExRateService != nil, func(ctx context.Context) {
		services.ExRateService.Run(ctx)
	})
	add("webhook", services.WebHookService != nil, func(ctx context.Context) {
		services.WebHookService.Run(ctx)
	})
	add("withdraw", services.WithdrawService != nil, func(ctx context.Context) {
		services.WithdrawService.Run(ctx, models.AllBlockchain())
	})
	add("unconfirmed_collapser", services.UnconfirmedCollapser != nil, func(ctx context.Context) {
		services.UnconfirmedCollapser.Run(ctx, conf.Transactions.UnconfirmedCollapseInterval)
	})
	add("notification", services.NotificationService != nil, func(ctx context.Context) {
		services.NotificationService.Run(ctx)
	})
	add("processing_ping_monitor", true, func(ctx context.Context) {
		processingPingMonitor(ctx, services, l)
	})
	add("exchange", services.ExchangeService != nil, func(ctx context.Context) {
		services.ExchangeService.Run(ctx)
	})
	add("exchange_withdrawal_queue", services.ExchangeWithdrawalService != nil, func(ctx context.Context) {
		services.ExchangeWithdrawalService.RunWithdrawalQueue(ctx)
	})
	add("exchange_withdrawal_updater", services.ExchangeWithdrawalService != nil, func(ctx context.Context) {
		services.ExchangeWithdrawalService.RunWithdrawalUpdater(ctx)
	})
	add("exchange_rules", services.ExchangeRulesService != nil, func(ctx context.Context) {
		services.ExchangeRulesService.Run(ctx)
	})
	add("system_heartbeat", services.SystemService != nil, func(ctx context.Context) {
		services.SystemService.RunHeartbeatLoop(ctx)
	})
	add("balance_updater", services.BalanceUpdater != nil, func(ctx context.Context) {
		services.BalanceUpdater.Run(ctx, conf.Wallets.UpdateBalancesInterval)
	})
	add("wallet_balance_stats", services.WalletBalanceService != nil, func(ctx context.Context) {
		services.WalletBalanceService.ProcessingBalanceStatsInBackground(ctx, conf.Wallets.UpdateTronResourcesInterval)
	})
	add("aml_status_checker", services.AMLStatusChecker != nil, func(ctx context.Context) {
		services.AMLStatusChecker.Run(ctx)
	})

	return workers
}

func processingPingMonitor(ctx context.Context, services *service.Services, l logger.Logger) {
	tickTocker := time.NewTicker(5 * time.Second)
	defer tickTocker.Stop()
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
