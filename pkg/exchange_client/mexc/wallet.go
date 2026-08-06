package mexc

import "context"

type IMexcWallet interface {
	GetCoinsConfig(ctx context.Context)
}
