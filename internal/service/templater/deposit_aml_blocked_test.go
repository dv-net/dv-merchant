package templater

import (
	"testing"

	"github.com/cbroglie/mustache"
	mxlogger "github.com/dv-net/mx/logger"
	"github.com/jellydator/ttlcache/v3"
	"github.com/stretchr/testify/require"

	"github.com/dv-net/dv-merchant/internal/settings"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

func newTestService(t *testing.T) *Service {
	t.Helper()

	svc := &Service{
		logger:         logger.New("test", mxlogger.Config{Level: "fatal"}),
		templateCache:  ttlcache.New[string, *mustache.Template](),
		mailerSettings: &settings.MailerSettings{MailerSender: "noreply@example.com"},
	}
	require.NoError(t, svc.parsePartialsFS())
	require.NoError(t, svc.initializeLocalization())
	require.NoError(t, svc.preloadTemplates())

	return svc
}

func TestAssembleEmail_DepositAMLBlocked(t *testing.T) {
	svc := newTestService(t)

	payload := &DepositAMLBlockedPayload{
		BasePayload:          BasePayload{UserEmail: "payer@example.com", Language: "ru"},
		PaymentDate:          "2026-11-12 10:00:00",
		TokenAmount:          "100.5",
		TokenSymbol:          "USDT",
		UsdAmount:            "100.49",
		BlockchainName:       "TRON",
		BlockchainCurrencyID: "USDT.Tron",
		TransactionHash:      "0xabc123",
		ToAddress:            "TXyz",
		RiskLevel:            "high",
		RiskScore:            "87",
		BlockedTransactionID: "11111111-2222-3333-4444-555555555555",
		BlockReasons: []DepositAMLBlockReasonView{
			{IsTotalRiskScore: true, Value: "87"},
			{IsRiskLevel: true, Value: "3"},
			{IsSignal: true, SignalLabel: "Sanctions", Value: "28"},
			{IsSignal: true, Value: "5"},
		},
	}

	body, err := svc.AssembleEmail(payload)
	require.NoError(t, err)

	out := body.String()
	for _, want := range []string{
		"Subject: Ваш платёж заблокирован по результатам AML-проверки",
		"<title>Платёж заблокирован по результатам AML-проверки</title>",
		"Платёж заблокирован",
		"100.5 USDT",
		"0xabc123",
		"TXyz",
		"11111111-2222-3333-4444-555555555555",
		"Высокий",
		"Причины блокировки",
		"Общая оценка риска",
		"Sanctions (28)",
		"Другое (5)",
		"обратитесь в магазин",
	} {
		require.Contains(t, out, want)
	}
	require.NotContains(t, out, "{{")
}
