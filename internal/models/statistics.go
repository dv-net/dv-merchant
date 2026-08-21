package models

import "github.com/shopspring/decimal"

type AdminDashboardStatisticsDTO struct {
	UsersCount       int64
	ProjectsCount    int64
	TurnoverTodayUSD decimal.Decimal
}

type AMLStatisticsDTO struct {
	CheckedToday    int64
	SuccessfulToday int64
	FailedToday     int64
}
