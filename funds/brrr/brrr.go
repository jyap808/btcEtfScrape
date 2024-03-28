package brrr

import (
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jyap808/btcEtfScrape/types"
)

func Collect() (result types.Result) {
	// Create a new collector
	c := colly.NewCollector()

	// Find and visit the target URL
	c.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		// Check the row and 1st column text
		if strings.Contains(e.ChildText("td:nth-of-type(1)"), "XBTUSD") {
			// Extract
			totalBitcoinRaw := e.ChildText("td:nth-of-type(4)")
			inputClean := strings.ReplaceAll(totalBitcoinRaw, ",", "")
			totalBitcoinInTrust, _ := strconv.ParseFloat(inputClean, 64)
			result.TotalBitcoin = totalBitcoinInTrust
			return
		}
	})

	// Visit the website
	c.Visit("https://valkyrieinvest.com/brrr-holdings/")

	c.Wait()

	return result
}
