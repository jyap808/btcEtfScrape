package bitb

import (
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/jyap808/btcEtfScrape/types"
)

func Collect() (result types.Result) {
	// Create a new collector
	c := colly.NewCollector()

	// Find and visit the target URL
	c.OnHTML("div[class*='layout-base']", func(e *colly.HTMLElement) {
		// Check if the div contains the desired text
		if strings.Contains(e.Text, "Bitcoin in Trust") {
			// Look for the div containing the value
			e.ForEach("div", func(_ int, el *colly.HTMLElement) {
				if strings.Contains(el.Text, "Bitcoin in Trust") {
					// Get the next div element which contains the figure
					figure := el.DOM.Next().Text()
					// Print the figure
					inputClean := strings.ReplaceAll(figure, ",", "")
					totalBitcoinInTrust, _ := strconv.ParseFloat(inputClean, 64)
					result.TotalBitcoin = totalBitcoinInTrust
					return
				}
			})
		}
	})

	// Visit the website
	c.Visit("https://bitbetf.com")

	c.Wait()

	return result
}
