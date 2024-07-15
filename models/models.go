package models

import "time"

type Finding struct {
	HistDate    time.Time `gorm:"column:hist_date"`
	TradingName string    `gorm:"column:trading_name"`
	CurrValue   float64   `gorm:"column:curr_value"`
	TrendValue  float64   `gorm:"column:trend_value"`
	TargetPerc  float64   `gorm:"column:target_perc"`
	MaxValue    float64   `gorm:"column:max_value"`
	MinValue    float64   `gorm:"column:min_value"`
	Status      string    `gorm:"column:status"`
	CloseValue  float64   `gorm:"column:close_value"`
	CloseDate   time.Time `gorm:"column:close_date"`
}

type TradingValue struct {
	HistDate    time.Time `gorm:"column:hist_date"`
	TradingName string    `gorm:"column:trading_name"`
	CurrValue   float64   `gorm:"column:curr_value"`
}

type RankingItem struct {
	TradingName string
	Trend       string
	PercWin     float64
	TotalWin    int
	Total       int
}

func (Finding) TableName() string      { return "findings_history" }
func (TradingValue) TableName() string { return "hist_trading_values" }
