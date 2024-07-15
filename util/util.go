package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"validar/config"
	"validar/models"
)

func SendMessageToDiscord(message, url string) error {
	config.ReadFile()
	if url != "" {
		url := url
		method := "POST"

		payload := strings.NewReader(fmt.Sprintf(`{
        	"content": "%s"
    	}`, message))

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return nil
	}
	return nil
}

func SendEmbedToDiscord(rankings []models.RankingItem, url string) error {
	if url != "" {
		var fields []map[string]interface{}
		for _, item := range rankings {
			field := map[string]interface{}{
				"name":   fmt.Sprintf(item.TradingName + " - " + item.Trend),
				"value":  fmt.Sprintf("Porcentagem: **%.2f**\nW/L: **%d/%d**", item.PercWin, item.TotalWin, item.Total-item.TotalWin),
				"inline": false,
			}
			fields = append(fields, field)
		}

		embed := map[string]interface{}{
			"title":       "Ranking Ativos 70%",
			"description": "Top 70% últimos 7 dias.",
			"color":       5814783,
			"fields":      fields,
			"footer": map[string]interface{}{
				"text": "Atualizado em",
			},
			"timestamp": time.Now().Format(time.RFC3339),
		}

		message := map[string]interface{}{
			"embeds": []map[string]interface{}{embed},
		}

		jsonData, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("Erro ao Marshal Json: %v", err)
		}

		client := &http.Client{}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("Erro ao criar nova requisição: %v", err)
		}
		req.Header.Add("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Erro ao enviar requisição: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			body, _ := ioutil.ReadAll(res.Body)
			return fmt.Errorf("Status Code Inesperado: %d, response: %s", res.StatusCode, string(body))
		}

		return nil
	}
	return nil
}

func GetNextCycleTime(now time.Time) time.Time {
	cycles := []time.Time{
		time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, now.Location()),
		time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location()),
		time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location()),
	}

	for i, cycle := range cycles {
		if now.Before(cycle) {
			return cycles[i]
		}
	}
	return cycles[0].Add(24 * time.Hour)
}
