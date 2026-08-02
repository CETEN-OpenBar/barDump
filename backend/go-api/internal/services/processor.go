package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type M map[string]interface{}

func parseMongoDBValue(val interface{}) interface{} {
	m, ok := val.(map[string]interface{})
	if !ok {
		return val
	}
	if oid, ok := m["$oid"]; ok {
		return oid
	}
	if nLong, ok := m["$numberLong"]; ok {
		if s, ok := nLong.(string); ok {
			n, _ := strconv.ParseInt(s, 10, 64)
			return n / 1000
		}
	}
	if binary, ok := m["$binary"].(map[string]interface{}); ok {
		if b64, ok := binary["base64"]; ok {
			return b64
		}
	}
	if data, ok := m["Data"].(string); ok {
		if _, hasSubtype := m["Subtype"]; hasSubtype {
			return data
		}
	}
	return val
}

func accountIDToBase64(accountID string) string {
	u, err := uuid.Parse(accountID)
	if err != nil {
		return ""
	}
	b, _ := u.MarshalBinary()
	return base64.StdEncoding.EncodeToString(b)
}

type UserStats struct {
	AccountID                  string
	AccountName                string
	Email                      string
	TotalDepense               int
	Articles                   map[string]int
	ArticlesParMois            map[string]int
	ArticlesParMoisDetail      map[string]map[string]int
	NombreTransactionRemote    int
	NombreTransactionNonRemote int
}

func ProcessTransactions(inputFile, outputFile, accountsFile, startDateStr, endDateStr string) error {
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	}
	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	}

	// Read accounts mapping
	emailMapping := make(map[string]string)
	if accountsFile != "" {
		accFile, err := os.Open(accountsFile)
		if err == nil {
			decoder := json.NewDecoder(accFile)
			if t, err := decoder.Token(); err == nil {
				if d, ok := t.(json.Delim); ok && d == '[' {
					for decoder.More() {
						var acc M
						if err := decoder.Decode(&acc); err == nil {
							accIDVal := parseMongoDBValue(acc["id"])
							emailVal := acc["email_address"]

							accID, ok1 := accIDVal.(string)
							email, ok2 := emailVal.(string)

							if ok1 && ok2 && accID != "" && email != "" {
								emailMapping[accID] = email
							}
						}
					}
				}
			}
			accFile.Close()
		}
	}

	userStats := make(map[string]*UserStats)

	txFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("open transactions: %w", err)
	}
	defer txFile.Close()

	txDecoder := json.NewDecoder(txFile)
	if t, err := txDecoder.Token(); err != nil {
		return fmt.Errorf("read transactions token: %w", err)
	} else if d, ok := t.(json.Delim); !ok || d != '[' {
		return fmt.Errorf("expected opening bracket")
	}

	for txDecoder.More() {
		var tx M
		if err := txDecoder.Decode(&tx); err != nil {
			return fmt.Errorf("decode transaction: %w", err)
		}
		if state, _ := tx["state"].(string); state != "finished" {
			continue
		}

		createdVal := parseMongoDBValue(tx["created_at"])
		var ts int64
		if f, ok := createdVal.(float64); ok {
			ts = int64(f)
		} else if i, ok := createdVal.(int64); ok {
			ts = i
		}

		date := time.Unix(ts, 0)
		if !startDate.IsZero() && date.Before(startDate) {
			continue
		}
		if !endDate.IsZero() && date.After(endDate) {
			continue
		}

		accID, ok1 := tx["account_id"].(string)
		accName, ok2 := tx["account_name"].(string)
		if !ok1 || !ok2 || accID == "" || accName == "" {
			continue
		}

		userKey := accID + "|" + accName
		stats, ok := userStats[userKey]
		if !ok {
			stats = &UserStats{
				AccountID:             accID,
				AccountName:           accName,
				Articles:              make(map[string]int),
				ArticlesParMois:       make(map[string]int),
				ArticlesParMoisDetail: make(map[string]map[string]int),
			}
			userStats[userKey] = stats
		}

		if isRemote, _ := tx["is_remote"].(bool); isRemote {
			stats.NombreTransactionRemote++
		} else {
			stats.NombreTransactionNonRemote++
		}

		monthKey := date.Format("2006-01")

		items, ok := tx["items"].([]interface{})
		if !ok {
			continue
		}

		for _, itemIntf := range items {
			item, ok := itemIntf.(map[string]interface{})
			if !ok || item["state"] == "canceled" {
				continue
			}

			itemName, _ := item["item_name"].(string)

			var itemAmount int
			amountVal := parseMongoDBValue(item["item_amount"])
			if f, ok := amountVal.(float64); ok {
				itemAmount = int(f)
			}

			var totalCost int
			costVal := parseMongoDBValue(item["total_cost"])
			if f, ok := costVal.(float64); ok {
				totalCost = int(f)
			}

			if itemName != "" && totalCost > 0 {
				stats.Articles[itemName] += itemAmount
				stats.TotalDepense += totalCost
				stats.ArticlesParMois[monthKey] += itemAmount

				if stats.ArticlesParMoisDetail[monthKey] == nil {
					stats.ArticlesParMoisDetail[monthKey] = make(map[string]int)
				}
				stats.ArticlesParMoisDetail[monthKey][itemName] += itemAmount
			}
		}
	}

	// top consumers
	productConsumers := make(map[string]map[string]int)
	for userKey, stats := range userStats {
		for article, qty := range stats.Articles {
			if productConsumers[article] == nil {
				productConsumers[article] = make(map[string]int)
			}
			productConsumers[article][userKey] = qty
		}
	}

	type topConsumerInfo struct {
		AccountID     string
		AccountName   string
		Quantite      int
		TotalQuantity int
	}

	topConsumersPerProduct := make(map[string]topConsumerInfo)
	for product, consumers := range productConsumers {
		type kv struct {
			Key   string
			Value int
		}
		var sorted []kv
		totalQuantity := 0
		for k, v := range consumers {
			sorted = append(sorted, kv{k, v})
			totalQuantity += v
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Value > sorted[j].Value
		})

		topUser := sorted[0]
		if topUser.Value > 5 {
			stats := userStats[topUser.Key]
			topConsumersPerProduct[product] = topConsumerInfo{
				AccountID:     stats.AccountID,
				AccountName:   stats.AccountName,
				Quantite:      topUser.Value,
				TotalQuantity: totalQuantity,
			}
		}
	}

	userTopProduct := make(map[string]M)
	for product, info := range topConsumersPerProduct {
		userKey := info.AccountID + "|" + info.AccountName
		existing, ok := userTopProduct[userKey]
		if !ok || info.TotalQuantity > existing["total"].(int) {
			userTopProduct[userKey] = M{
				"nom":      product,
				"quantite": info.Quantite,
				"total":    info.TotalQuantity,
			}
		}
	}

	// Final Result List
	var resultList []M

	for userKey, stats := range userStats {
		// top 3
		type kv struct {
			Key   string
			Value int
		}
		var sortedArticles []kv
		for k, v := range stats.Articles {
			sortedArticles = append(sortedArticles, kv{k, v})
		}
		sort.Slice(sortedArticles, func(i, j int) bool {
			return sortedArticles[i].Value > sortedArticles[j].Value
		})
		var top3 []M
		for i := 0; i < len(sortedArticles) && i < 3; i++ {
			top3 = append(top3, M{"nom": sortedArticles[i].Key, "quantite": sortedArticles[i].Value})
		}

		articlesParMoisInfo := make(M)
		for monthKey, count := range stats.ArticlesParMois {
			details := stats.ArticlesParMoisDetail[monthKey]
			var topProd string
			maxCount := -1
			for p, c := range details {
				if c > maxCount {
					maxCount = c
					topProd = p
				}
			}

			var prodValue interface{}
			if topProd != "" {
				prodValue = topProd
			} else {
				prodValue = nil
			}

			articlesParMoisInfo[monthKey] = M{
				"nombre_transactions":    count,
				"produit_le_plus_achete": prodValue,
			}
		}

		accIDB64 := accountIDToBase64(stats.AccountID)

		resultList = append(resultList, M{
			"account_id":                    stats.AccountID,
			"account_name":                  stats.AccountName,
			"email":                         emailMapping[accIDB64],
			"total_depense":                 stats.TotalDepense,
			"top_3_produits":                top3,
			"nombre_produits_differents":    len(stats.Articles),
			"articles_par_mois":             articlesParMoisInfo,
			"top_consommateur_produit":      userTopProduct[userKey],
			"nombre_transaction_remote":     stats.NombreTransactionRemote,
			"nombre_transaction_non_remote": stats.NombreTransactionNonRemote,
		})
	}

	// Sorting and ranking
	sort.Slice(resultList, func(i, j int) bool {
		return resultList[i]["total_depense"].(int) > resultList[j]["total_depense"].(int)
	})

	totalUsers := len(resultList)
	for i := range resultList {
		rankPercentage := float64(i+1) / float64(totalUsers) * 100
		resultList[i]["rang"] = i + 1
		resultList[i]["rang_pourcentage"] = float64(int(rankPercentage*100)) / 100
	}

	sort.Slice(resultList, func(i, j int) bool {
		return resultList[i]["nombre_produits_differents"].(int) > resultList[j]["nombre_produits_differents"].(int)
	})
	for i := range resultList {
		rankPercentage := float64(i+1) / float64(totalUsers) * 100
		resultList[i]["rang_produits_differents"] = i + 1
		resultList[i]["rang_pourcentage_produits_differents"] = float64(int(rankPercentage*100)) / 100
	}

	sort.Slice(resultList, func(i, j int) bool {
		return resultList[i]["total_depense"].(int) > resultList[j]["total_depense"].(int)
	})

	finalResult := M{"utilisateurs": resultList}

	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(finalResult); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	log.Printf("Traitement terminé: %d utilisateurs", totalUsers)
	return nil
}
