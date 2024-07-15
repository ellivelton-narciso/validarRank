package main

import (
	"fmt"
	"log"
	"time"
	"validar/config"
	"validar/database"
	"validar/models"
	"validar/util"
)

func fetchFindings() ([]models.Finding, error) {
	var findings []models.Finding
	result := database.DB.Where("status = ?", "R").Find(&findings)
	return findings, result.Error
}

func fetchWinningFindings() ([]models.Finding, error) {
	var findings []models.Finding
	result := database.DB.Where("status = ? AND hist_date >= ?", "W", time.Now().Add(-4*time.Hour)).Find(&findings)
	return findings, result.Error
}

func fetchCurrentPrice(tradingName string) (float64, error) {
	var tradingValue models.TradingValue
	result := database.DB.Where("trading_name = ?", tradingName).Order("hist_date desc").First(&tradingValue)
	return tradingValue.CurrValue, result.Error
}

func fetchPriceAtTime(tradingName string, histDate time.Time) (float64, error) {
	var tradingValue models.TradingValue
	result := database.DB.Where("trading_name = ? AND hist_date <= ?", tradingName, histDate).Order("hist_date desc").First(&tradingValue)
	return tradingValue.CurrValue, result.Error
}

func updateFindingStatus(histDate time.Time, tradingName string, status string, closeValue float64, closeDate time.Time) error {
	result := database.DB.Model(&models.Finding{}).Where("hist_date = ? AND trading_name = ?", histDate, tradingName).Updates(map[string]interface{}{
		"status":      status,
		"close_value": closeValue,
		"close_date":  closeDate,
	})
	return result.Error
}

func updateFindingValues(histDate time.Time, tradingName string, maxValue, minValue float64) error {
	result := database.DB.Model(&models.Finding{}).Where("hist_date = ? AND trading_name = ?", histDate, tradingName).Updates(map[string]interface{}{
		"max_value": maxValue,
		"min_value": minValue,
	})
	return result.Error
}

func processFindings() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			findings, err := fetchFindings()
			if err != nil {
				log.Println("Erro ao buscar findings:", err)
				continue
			}

			for _, finding := range findings {
				currentPrice, err := fetchCurrentPrice(finding.TradingName)
				if err != nil {
					log.Println("Erro ao buscar preço atual:", err)
					continue
				}

				newMaxValue := finding.MaxValue
				newMinValue := finding.MinValue

				if currentPrice > finding.MaxValue {
					newMaxValue = currentPrice
				}
				if currentPrice < finding.MinValue || finding.MinValue == 0 {
					newMinValue = currentPrice
				}

				err = updateFindingValues(finding.HistDate, finding.TradingName, newMaxValue, newMinValue)
				if err != nil {
					log.Println("Erro ao atualizar max/min values:", err)
					continue
				}

				var status string
				status = ""

				if finding.TrendValue == 1 { // LONG
					if currentPrice >= finding.CurrValue*(1+(finding.TargetPerc/100)) {
						status = "W"
						err := util.SendMessageToDiscord("["+finding.TradingName+"] LONG - Ganhou. "+fmt.Sprintf("%.6f", currentPrice), config.AlertasDisc)
						if err != nil {
							log.Println("Erro ao enviar mensagem pro discord")
						}

					} else if currentPrice <= finding.CurrValue*(1-(finding.TargetPerc/100)) {
						status = "L"
						err := util.SendMessageToDiscord("["+finding.TradingName+"] LONG - Perdeu. "+fmt.Sprintf("%.6f", currentPrice), config.AlertasDisc)
						if err != nil {
							log.Println("Erro ao enviar mensagem pro discord")
						}
					}
				}

				if finding.TrendValue == -1 { // SHORT
					if currentPrice <= finding.CurrValue*(1-(finding.TargetPerc/100)) {
						status = "W"
						err := util.SendMessageToDiscord("["+finding.TradingName+"] SHORT - Ganhou. "+fmt.Sprintf("%.6f", currentPrice), config.AlertasDisc)
						if err != nil {
							log.Println("Erro ao enviar mensagem pro discord")
						}
					} else if currentPrice >= finding.CurrValue*(1+(finding.TargetPerc/100)) {
						status = "L"
						err := util.SendMessageToDiscord("["+finding.TradingName+"] SHORT - Perdeu. "+fmt.Sprintf("%.6f", currentPrice), config.AlertasDisc)
						if err != nil {
							log.Println("Erro ao enviar mensagem pro discord")
						}
					}
				}

				if status != "" {
					err = updateFindingStatus(finding.HistDate, finding.TradingName, status, currentPrice, time.Now())
					if err != nil {
						log.Println("Error updating status:", err)
					}
				}
			}
		}
	}
}

func monitorWinningFindings() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			findings, err := fetchWinningFindings()
			if err != nil {
				log.Println("Erro ao buscar findings ganhadores:", err)
				continue
			}

			for _, finding := range findings {
				currentPrice, err := fetchCurrentPrice(finding.TradingName)
				if err != nil {
					log.Println("Erro ao buscar preço atual:", err)
					continue
				}

				previousPrice, err := fetchPriceAtTime(finding.TradingName, time.Now().Add(-5*time.Minute))
				if err != nil {
					log.Println("Erro ao buscar preço de 5 minutos atrás:", err)
					continue
				}

				if finding.TrendValue == 1 { // LONG
					if currentPrice > finding.MaxValue {
						if currentPrice > previousPrice {
							newMaxValue := currentPrice
							err = updateFindingValues(finding.HistDate, finding.TradingName, newMaxValue, finding.MinValue)
							if err != nil {
								log.Println("Erro ao atualizar max value:", err)
								continue
							}
							err = updateFindingStatus(finding.HistDate, finding.TradingName, "W", currentPrice, time.Now())
							if err != nil {
								log.Println("Erro ao atualizar status:", err)
							}
						}
					}
				} else if finding.TrendValue == -1 { // SHORT
					if currentPrice < finding.MinValue {
						if currentPrice < previousPrice {
							newMinValue := currentPrice
							err = updateFindingValues(finding.HistDate, finding.TradingName, finding.MaxValue, newMinValue)
							if err != nil {
								log.Println("Erro ao atualizar min value:", err)
								continue
							}
							err = updateFindingStatus(finding.HistDate, finding.TradingName, "W", currentPrice, time.Now())
							if err != nil {
								log.Println("Erro ao atualizar status:", err)
							}
						}
					}
				}
			}
		}
	}
}

func sendRanking() {
	var rankings []models.RankingItem

	query := `
		-- TOP 10 ALERTAS COM WIN RATE 70% OU MAIS
		SELECT trading_name,
			   trend,
			   ROUND(total_win / total * 100, 2) perc_win,
			   total_win,
			   total
		FROM (
			SELECT TIPO_ALERTA,
			trend,
			trading_name,
			SUM(CASE WHEN status = 'W' THEN 1 ELSE 0 END) AS total_win,
			COUNT(1)                                      AS total
			FROM (
				SELECT  ROUND(other_value)                                       AS TIPO_ALERTA,
						trading_name,
						(CASE WHEN trend_value > 0 THEN 'LONG' ELSE 'SHORT' END) AS trend,
						status
				FROM findings_history a
				WHERE close_date > NOW() - INTERVAL 7 DAY
				  AND status IN ('W', 'L')
			  ) x
			  GROUP BY TIPO_ALERTA, trading_name, trend) z
			WHERE z.TIPO_ALERTA IN (31)
			AND total > 1
			AND ROUND(total_win / total * 100, 2) >= 70
		ORDER BY perc_win DESC, total_win DESC;
	`

	if err := database.DB.Raw(query).Scan(&rankings).Error; err != nil {
		log.Printf("Erro ao executar query: %v\n", err)
		return
	}

	if len(rankings) == 0 {
		return
	}
	urlRanking := "https://discord.com/api/webhooks/1246612954074452088/5i5R6QuCcclbSzCPOhciKbYIbU3q622wD24Rnf1Ouj6FxYYA_g2GEQqrwiSEpThoZ07b"
	if err := util.SendEmbedToDiscord(rankings, urlRanking); err != nil {
		log.Printf("Erro ao enviar ranking pro discord: %v\n", err)
		return
	}
}

func checkAndSendRankings() {
	sendRanking()

	now := time.Now()
	nextCycle := util.GetNextCycleTime(now)
	timeUntilNextCycle := time.Until(nextCycle)

	time.Sleep(timeUntilNextCycle)
	sendRanking()
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sendRanking()
		}
	}
}

func main() {
	database.DBCon()

	go processFindings()
	go monitorWinningFindings()
	go checkAndSendRankings()

	select {}
}
