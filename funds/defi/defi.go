package defi

import (
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jyap808/btcEtfScrape/types"
)

func Collect() (result types.Result) {
	url := "https://hashdex-etfs.com/defi"

	c := colly.NewCollector()

	// Find and visit the table
	c.OnHTML("table.table-holdings", func(e *colly.HTMLElement) {
		// Iterate over each row in the table
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			// Check if the row contains the target value
			if strings.Contains(row.Text, "BITCOIN") {
				// Find the cell containing the value
				totalBitcoinInTrustRaw := row.ChildText("td.shares-holding")
				inputClean := strings.ReplaceAll(totalBitcoinInTrustRaw, ",", "")
				inputClean = strings.TrimSpace(inputClean)
				totalBitcoinInTrust, _ := strconv.ParseFloat(inputClean, 64)

				result.TotalBitcoin = totalBitcoinInTrust
			}
		})
	})

	// Visit the URL
	c.Visit(url)

	return result
}
