package gbtc

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type nextData struct {
	Props struct {
		PageProps struct {
			Page struct {
				Includes map[string]interface{}
			}
		}
	}
}

type Result struct {
	TotalBitcoinInTrust float64
	Date                time.Time
}

func Collect() (result Result) {
	// creating a new Colly instance
	c := colly.NewCollector()

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	// Set up a callback to be executed when the HTML body is found
	c.OnHTML("body", func(e *colly.HTMLElement) {
		// Get the content of the __NEXT_DATA__ script tag
		nextDataContent := e.DOM.Find("#__NEXT_DATA__").Text()

		// Parse the content as JSON
		var data nextData //[string]interface{}
		err := json.NewDecoder(strings.NewReader(nextDataContent)).Decode(&data)
		if err != nil {
			log.Fatal(err)
		}

		// Access the "includes" field
		includesData := data.Props.PageProps.Page.Includes

		// Search for the value containing "totalBitcoinInTrust" within "includes"
		result, err = findResultsInIncludes(includesData)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Set up error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// visiting the target page
	c.Visit("https://etfs.grayscale.com/gbtc")

	c.Wait()

	log.Printf("Scraping finished")

	return result
}

// findResultsInIncludes searches for the unique field within "includes"
func findResultsInIncludes(includesData map[string]interface{}) (Result, error) {
	result := &Result{}

	for _, value := range includesData {
		// Assuming the value is a map[string]interface{}
		include, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		// Search for "totalBitcoinInTrustRaw" within each include
		totalBitcoinInTrustRaw, found := include["totalBitcoinInTrust"].(string)
		if found {
			inputClean := strings.ReplaceAll(totalBitcoinInTrustRaw, ",", "")
			totalBitcoinInTrust, _ := strconv.ParseFloat(inputClean, 64)

			// Define the layout of the input date
			layout := "01/02/2006"
			// Parse the string as a time.Time value
			parsedTime, _ := time.Parse(layout, include["date"].(string))

			result.TotalBitcoinInTrust = totalBitcoinInTrust
			result.Date = parsedTime
		}

		return *result, nil
	}

	return Result{}, fmt.Errorf("totalBitcoinInTrust not found within 'includes'")
}
