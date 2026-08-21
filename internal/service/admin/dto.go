package admin

import "github.com/shopspring/decimal"

type DashboardStatisticsDTO struct {
	UsersCount       int64
	ProjectsCount    int64
	TurnoverTodayUSD decimal.Decimal
}
