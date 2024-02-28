package ibit

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jyap808/btcEtfScrape/types"
)

type FundData struct {
	AaData [][]interface{} `json:"aaData"`
}

type Fund struct {
	Ticker string
	Shares types.Result
}

func Collect() (result types.Result) {
	url := "https://blackrock.com/us/financial-professionals/products/333011/fund/1500962885783.ajax?tab=all&fileType=json"

	// Create a new HTTP client
	client := http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error performing request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return
	}

	// Trim any leading characters that may cause the issue
	bodyStr := string(body)
	bodyStr = strings.TrimLeftFunc(bodyStr, func(r rune) bool {
		return r != '{' && r != '['
	})

	var data FundData
	if err := json.Unmarshal([]byte(bodyStr), &data); err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
	}

	// Iterate through the funds and find the one with ticker "BTC"
	for _, fund := range data.AaData {
		if len(fund) > 0 && fund[0] == "BTC" {
			// Extract the "Shares" field
			sharesMap, ok := fund[6].(map[string]interface{})
			if ok {
				sharesRaw, _ := sharesMap["raw"].(float64)
				result.TotalBitcoin = sharesRaw
				return
			}
		}
	}

	return result
}
