package ezbc

import (
	"bytes"
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
	Portfolio Portfolio `json:"Portfolio"`
}

type Portfolio struct {
	PortfolioData PortfolioData `json:"portfolio"`
}

type PortfolioData struct {
	DailyHoldings []Holding `json:"dailyholdings"`
}

type Holding struct {
	Date     string `json:"asofdate"`
	SECName  string `json:"secname"`
	Quantity string `json:"quantityshrpar"`
}

func Collect() (result types.Result) {
	// This API key is hard coded on their web site
	url := "https://www.franklintempleton.com/api/pds/price-and-performance?apikey=4ef35821-5244-41bc-a699-0192d002c3d1p&op=Holdings&id=14"

	// JSON payload
	payload := []byte(`{
        "operationName": "Holdings",
        "variables": {
            "countrycode": "US",
            "languagecode": "en_US",
            "fundid": "39639"
        },
        "query": "query Holdings($fundid: String!, $countrycode: String!, $languagecode: String!) {Portfolio(fundid: $fundid countrycode: $countrycode languagecode: $languagecode) {portfolio {dailyholdings {asofdate secname quantityshrpar}}}}"
    }`)

	// Create a new HTTP client
	client := http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

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

	// Iterate
	for _, nav := range data.Data.Portfolio.PortfolioData.DailyHoldings {
		if nav.SECName == "BITCOIN" {
			// Extract
			totalRaw := nav.Quantity
			inputClean := strings.ReplaceAll(totalRaw, ",", "")
			total, _ := strconv.ParseFloat(inputClean, 64)
			result.TotalAsset = total

			// Define the layout of the input date
			layout := "01/02/2006"
			// Parse the string as a time.Time value
			parsedTime, _ := time.Parse(layout, nav.Date)
			result.Date = parsedTime
			return
		}
	}

	return result
}
