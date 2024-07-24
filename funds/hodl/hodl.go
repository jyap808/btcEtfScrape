package hodl

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jyap808/btcEtfScrape/types"
)

type FundData struct {
	Data Data `json:"data"`
}

type Data struct {
	Date string `json:"AsOfDate"`
	Navs []Nav
}

type Nav struct {
	Key   string
	Value string
}

func Collect() (result types.Result) {
	url := "https://www.vaneck.com/Main/NavInformationBlock/GetContent/?blockid=252190&ticker=HODL"

	// NOTE: Fix for getting old cached responses from this endpoint
	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

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

	var data FundData
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
	}

	// Define the layout of the input date
	layout := "01/02/2006"
	// Parse the string as a time.Time value
	parsedTime, _ := time.Parse(layout, data.Data.Date)
	result.Date = parsedTime

	// Iterate
	for _, nav := range data.Data.Navs {
		if nav.Key == "Bitcoin in Trust" {
			// Extract
			totalRaw := nav.Value
			inputClean := strings.ReplaceAll(totalRaw, ",", "")
			total, _ := strconv.ParseFloat(inputClean, 64)
			result.TotalAsset = total
			return
		}
	}

	return result
}
