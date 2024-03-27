package fbtc

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/jyap808/btcEtfScrape/types"

	"github.com/ledongthuc/pdf"
)

func Collect() (result types.Result) {
	actionExchangeRepositoryURL := getActionExchangeRepositoryURL()
	collectionID := getCollectionID(actionExchangeRepositoryURL)

	url := fmt.Sprintf("https://www.actionsxchangerepository.fidelity.com/ShowDocument/documentPDF.htm?clientId=Fidelity&applicationId=MFL&securityId=315948109&docType=DALY&docFormat=pdf&securityIdType=CUSIP&collectionId=%d&docName=1.WOB-DALY.pdf&criticalIndicator=N&pdfReaderStatus=Y", collectionID)

	// Fetch the data from the URL
	resp, err := http.Get(url)
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

	// Create a reader from the byte slice
	reader := bytes.NewReader(body)

	r, err := pdf.NewReader(reader, int64(len(body)))
	if err != nil {
		log.Println("Error creating reader:", err)
		return
	}

	pageIndex := 1
	p := r.Page(pageIndex)

	rows, _ := p.GetTextByRow()
	// hard code retrieve
	rowHoldings := rows[0].Content[9].S

	cols, _ := p.GetTextByColumn()
	// hard code retrieve
	colHoldings := cols[1].Content[23].S

	if rowHoldings == colHoldings {
		total, _ := strconv.ParseFloat(rowHoldings, 64)
		result.TotalBitcoin = total
	}

	return
}

func getActionExchangeRepositoryURL() (redirectURL string) {
	// This static URL redirects to www.actionsxchangerepository.fidelity.com
	url := "https://fundresearch.fidelity.com/prospectus/eproredirect?clientId=Fidelity&applicationId=MFL&securityIdType=CUSIP&critical=N&securityId=315948109"

	c := colly.NewCollector()

	// Find and extract the redirect URL
	c.OnHTML("a", func(e *colly.HTMLElement) {
		redirectURL = e.Attr("href")
	})

	// Visit the URL
	err := c.Visit(url)
	if err != nil {
		log.Println("Error:", err)
		return ""
	}

	return redirectURL
}

func getCollectionID(url string) (collectionID int) {
	c := colly.NewCollector()

	c.OnHTML("td", func(e *colly.HTMLElement) {
		// Check if the fundDocumentType is "DALY"
		if strings.Contains(e.Attr("onclick"), "'DALY'") {
			// Extract
			collectionID = extractCollectionIDFromOnClick(e.Attr("onclick"))
		}
	})

	// Visit the URL
	err := c.Visit(url)
	if err != nil {
		log.Println("Error:", err)
		return 0
	}

	return collectionID
}

func extractCollectionIDFromOnClick(onClick string) int {
	// Split the onClick attribute by comma and extract the element
	parts := strings.Split(onClick, ",")
	if len(parts) >= 5 {
		// Remove surrounding quotes and trim whitespace
		rawID := strings.TrimSpace(strings.Trim(parts[6], "'"))
		ID, _ := strconv.Atoi(rawID)
		return ID
	}
	return 0
}
