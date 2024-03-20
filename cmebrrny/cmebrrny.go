/*
The CME CF Bitcoin Reference Rate – New York Variant is a once a day (4pm ET)
benchmark price for bitcoin, measured in US dollars per bitcoin.

https://www.cmegroup.com/markets/cryptocurrencies/cme-cf-cryptocurrency-benchmarks.html
*/
package cmebrrny

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type ReferenceRate struct {
	Value float64   `json:",string"`
	Date  time.Time `json:"date"`
}

type ReferenceRates struct {
	BRRNY []ReferenceRate `json:"BRRNY"`
}

// Custom unmarshalling function for time.Time field
func (rr *ReferenceRate) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Value float64 `json:",string"`
		Date  string  `json:"date"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	date, err := time.ParseInLocation("2006-01-02 15:04:05", tmp.Date, time.UTC)
	if err != nil {
		return err
	}
	rr.Value = tmp.Value
	rr.Date = date
	return nil
}

// Return the latest CME BRR NY
func GetBRRYNY() (referenceRates []ReferenceRate, err error) {
	url := "https://www.cmegroup.com/services/cryptocurrencies/reference-rates"

	// Create a new HTTP client
	client := http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return []ReferenceRate{}, err
	}

	// Set headers
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2.1 Safari/605.1.15")

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error performing request:", err)
		return []ReferenceRate{}, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return []ReferenceRate{}, err
	}

	// Parse JSON data into struct
	var data map[string]ReferenceRates
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		return []ReferenceRate{}, err
	}

	return data["referenceRates"].BRRNY, nil
}
