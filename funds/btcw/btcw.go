package btcw

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jyap808/btcEtfScrape/types"
)

func Collect() (result types.Result) {
	// Instantiate a new collector
	c := colly.NewCollector()

	// Define the regular expressions to extract the date and sharespar
	dateRegex := regexp.MustCompile(`"COBDate":"(\d{2}/\d{2}/\d{4})"`)
	sharesParRegex := regexp.MustCompile(`"SharesPar":"(\d+(?:\.\d+)?)`)

	c.OnHTML("script", func(e *colly.HTMLElement) {
		// Check if the script contains the desired JavaScript snippet
		if strings.Contains(e.Text, "WTree.exporter.addExportedItem('current-day-holdings-table'") {
			// Extract the JavaScript code containing the data
			scriptText := e.Text

			// Check if the script contains the desired JavaScript snippet with BITCOIN
			if strings.Contains(scriptText, "BITCOIN") {
				// Extract the date using the dateRegex
				dateMatch := dateRegex.FindStringSubmatch(scriptText)
				if len(dateMatch) >= 2 {
					dateRaw := dateMatch[1]
					// Define the layout of the input date
					layout := "01/02/2006"
					// Parse the string as a time.Time value
					parsedTime, _ := time.Parse(layout, dateRaw)
					result.Date = parsedTime
				}

				// Extract the sharespar using the sharesParRegex
				sharesParMatch := sharesParRegex.FindStringSubmatch(scriptText)
				if len(sharesParMatch) >= 2 {
					totalRaw := sharesParMatch[1]
					inputClean := strings.ReplaceAll(totalRaw, ",", "")
					total, _ := strconv.ParseFloat(inputClean, 64)
					result.TotalBitcoin = total
				}
			}
		}
	})

	c.Visit("https://www.wisdomtree.com/investments/global/etf-details/modals/all-current-day-holdings?id={E22BFB6E-98F0-4CAE-AFAE-699175D6F697}")

	c.Wait()

	return result
}
