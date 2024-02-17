package ibit

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

type FundData struct {
	AaData [][]interface{} `json:"aaData"`
}

type Fund struct {
	Ticker string
	Shares Shares
}

type Shares struct {
	Display string  `json:"display"`
	Raw     float64 `json:"raw"`
}

func Collect() float64 {
	url := "https://blackrock.com/us/financial-professionals/products/333011/fund/1500962885783.ajax?tab=all&fileType=json"

	// Create a new HTTP client
	client := http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return 0
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error performing request:", err)
		return 0
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return 0
	}

	// Trim any leading characters that may cause the issue
	bodyStr := string(body)
	bodyStr = strings.TrimLeftFunc(bodyStr, func(r rune) bool {
		return r != '{' && r != '['
	})

	var data FundData
	if err := json.Unmarshal([]byte(bodyStr), &data); err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Iterate through the funds and find the one with ticker "BTC"
	var btcShares Shares
	for _, fund := range data.AaData {
		if len(fund) > 0 && fund[0] == "BTC" {
			// Extract the "Shares" field
			sharesMap, ok := fund[6].(map[string]interface{})
			if ok {
				sharesDisplay, _ := sharesMap["display"].(string)
				sharesRaw, _ := sharesMap["raw"].(float64)
				btcShares = Shares{Display: sharesDisplay, Raw: sharesRaw}
				break
			}
		}
	}

	return btcShares.Raw
}
