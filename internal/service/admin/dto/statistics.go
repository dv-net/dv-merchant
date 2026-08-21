package dto

import "github.com/shopspring/decimal"

type DashboardStatistics struct {
	UsersCount       int64
	ProjectsCount    int64
	TurnoverTodayUSD decimal.Decimal
}
