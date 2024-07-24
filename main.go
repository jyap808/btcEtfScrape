package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/jyap808/btcEtfScrape/cmebrrny"
	"github.com/jyap808/btcEtfScrape/funds"
	"github.com/jyap808/btcEtfScrape/types"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	gotwiTypes "github.com/michimani/gotwi/tweet/managetweet/types"
)

type payload struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Embeds    []embed `json:"embeds"`
}

type embed struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type manualData struct {
	Ticker string
	Result types.Result
}

// Hold details for each ticker
type tickerDetail struct {
	Description string
	Note        string
	Delayed     bool
}

var (
	webhookURL string

	avatarUsername string
	avatarURL      string

	// track
	tickerResults         = map[string]types.Result{}
	tickerResultsOverride = map[string]types.Result{}
	asset_rr              [5]cmebrrny.ReferenceRate

	// polling intervals
	pollMinutes  int = 5
	backoffHours int = 12

	// Skip X post when the difference is under this threshold
	minBitcoinDiff float64 = 1.0

	tickerDetails = map[string]tickerDetail{
		"ARKB": {Description: "Ark 21Shares", Note: "ARKB holdings are usually updated 10+ hours after the close of trading"}, // ARK 21Shares Bitcoin ETF
		"BITB": {Description: "Bitwise", Note: "BITB holdings are usually updated 4.5+ hours after the close of trading"},     // Bitwise Bitcoin ETF
		"BRRR": {Description: "Valkyrie", Note: "BRRR holdings are usually updated 10+ hours after the close of trading"},     // Valkyrie Bitcoin Fund
		"BTCW": {Description: "WisdomTree", Note: ""},                                                                         // WisdomTree Bitcoin Fund
		"DEFI": {Description: "Hashdex", Note: ""},                                                                            // Hashdex Bitcoin ETF
		"EZBC": {Description: "Franklin", Note: "EZBC holdings are usually updated 5.5+ hours after the close of trading"},    // Franklin Bitcoin ETF
		"FBTC": {Description: "Fidelity", Note: "FBTC holdings are usually updated 16+ hours after the close of trading"},     // Fidelity Wise Origin Bitcoin Fund
		"GBTC": {Description: "Grayscale", Note: "GBTC holdings are usually updated 1 day late", Delayed: true},               // Grayscale Bitcoin Trust
		"HODL": {Description: "VanEck", Note: "HODL holdings are usually updated 1 day late", Delayed: true},                  // VanEck Bitcoin Trust
		"IBIT": {Description: "BlackRock", Note: "IBIT holdings are usually updated 13+ hours after the close of trading"},    // iShares Bitcoin Trust
	}
	// BTCO - Invesco Galaxy Bitcoin ETF
)

const (
	OAuthTokenEnvKeyName       = "GOTWI_ACCESS_TOKEN"
	OAuthTokenSecretEnvKeyName = "GOTWI_ACCESS_TOKEN_SECRET"
)

func init() {
	flag.StringVar(&webhookURL, "webhookURL", "https://discord.com/api/webhooks/", "Webhook URL")
	flag.StringVar(&avatarUsername, "avatarUsername", "Annalee Call", "Avatar username")
	flag.StringVar(&avatarURL, "avatarURL", "https://static1.personality-database.com/profile_images/6604632de9954b4d99575e56404bd8b7.png", "Avatar image URL")
	flag.Parse()
}

func main() {
	// Initialize empty tickerResult
	wgCount := 0
	for ticker := range tickerDetails {
		tickerResults[ticker] = types.Result{}
		tickerResultsOverride[ticker] = types.Result{}
		wgCount++
	}

	// Initialize cmebrrnyRR
	asset_rr = getCMEBRRNYRR()
	if asset_rr[0].Value == 0 {
		log.Fatalln("Error: CME BRR NY initialization error")
	}

	var wg sync.WaitGroup

	// Increment the WaitGroup counter for each scraping function
	wg.Add(wgCount)

	// Launch goroutines for scraping functions
	go handleFund(&wg, funds.ArkbCollect, "ARKB")
	go handleFund(&wg, funds.BitbCollect, "BITB")
	go handleFund(&wg, funds.BrrrCollect, "BRRR")
	go handleFund(&wg, funds.BtcwCollect, "BTCW")
	go handleFund(&wg, funds.DefiCollect, "DEFI")
	go handleFund(&wg, funds.EzbcCollect, "EZBC")
	go handleFund(&wg, funds.FbtcCollect, "FBTC")
	go handleFund(&wg, funds.GbtcCollect, "GBTC")
	go handleFund(&wg, funds.HodlCollect, "HODL")
	go handleFund(&wg, funds.IbitCollect, "IBIT")

	// Manual endpoints
	http.HandleFunc("/override", handleOverride)
	http.HandleFunc("/update", handleUpdate)

	// Start HTTP server in a separate goroutine
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	log.Println("All scraping functions have finished.")
}

// Generic handler
func handleFund(wg *sync.WaitGroup, collector func() types.Result, ticker string) {
	defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes

	for {
		var newResult types.Result
		override := false

		// Check if there is a manual override set
		if tickerResultsOverride[ticker].TotalAsset != 0 {
			newResult = tickerResultsOverride[ticker]

			// Clear override
			tickerResultsOverride[ticker] = types.Result{}
			override = true
		} else {
			newResult = collector()
		}

		// Check date is valid. Date is optional so we check it is not none
		if !newResult.Date.IsZero() && newResult.Date.Before(tickerResults[ticker].Date) {
			log.Printf("%s new result before current: %+v", ticker, newResult)

			// Backoff for 1 hr or this just will loop
			time.Sleep(time.Hour * time.Duration(1))

			continue
		}

		if newResult.TotalAsset != tickerResults[ticker].TotalAsset && newResult.TotalAsset != 0 {
			if tickerResults[ticker].TotalAsset == 0 {
				// initialize
				tickerResults[ticker] = newResult
				log.Printf("Initialize %s: %+v", ticker, tickerResults[ticker])
			} else {
				// compare
				assetDiff := newResult.TotalAsset - tickerResults[ticker].TotalAsset
				rr := getCMEBRRNYRR()
				assetPrice := rr[0].Value
				if tickerDetails[ticker].Delayed {
					assetPrice = rr[1].Value
				}
				flowDiff := assetDiff * assetPrice

				header := ticker
				if newResult.Date != (time.Time{}) {
					layout := "01/02/2006"
					formattedTime := newResult.Date.Format(layout)

					header = fmt.Sprintf("%s %s", ticker, formattedTime)
				}

				msg := fmt.Sprintf("%s\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS Flow: $%.1f, CMEBRRNY: $%.1f",
					header, assetDiff, newResult.TotalAsset,
					flowDiff, rr[0].Value)

				postDiscord(msg)

				flowEmoji := "🚀"
				if assetDiff < 0 {
					flowEmoji = "👎"
				}

				note := ""
				if !override {
					note = tickerDetails[ticker].Note
				}

				xMsg := fmt.Sprintf("%s $%s\n\n%s FLOW: %s BTC, $%s\n🏦 TOTAL Bitcoin in Trust: %s $BTC\n\n%s",
					tickerDetails[ticker].Description, ticker,
					flowEmoji, humanize.CommafWithDigits(assetDiff, 2), humanize.CommafWithDigits(flowDiff, 0),
					humanize.CommafWithDigits(newResult.TotalAsset, 1), note)

				// Reporting threshold check. Get the absolute difference
				absAssetDiff := math.Abs(assetDiff)
				if absAssetDiff > minBitcoinDiff {
					postTweet(xMsg)
				}

				tickerResults[ticker] = newResult

				log.Printf("Update %s: %+v", ticker, tickerResults[ticker])

				time.Sleep(time.Hour * time.Duration(backoffHours))
			}
		}

		time.Sleep(time.Minute * time.Duration(pollMinutes))
	}
}

func handleData(w http.ResponseWriter, r *http.Request, updateType string) {
	var data manualData

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		return
	}

	// Set based on updateType
	var results map[string]types.Result
	switch updateType {
	case "override":
		results = tickerResultsOverride
	case "update":
		results = tickerResults
	default:
		log.Println("Invalid update type")
		return
	}

	// Update the corresponding map
	newResult := types.Result{TotalAsset: data.Result.TotalAsset, Date: data.Result.Date}
	results[data.Ticker] = newResult

	// Log and respond with success message
	log.Printf("Data %s %s: %+v", updateType, data.Ticker, results[data.Ticker])
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Data %s successful\n", updateType)
}

func handleOverride(w http.ResponseWriter, r *http.Request) {
	handleData(w, r, "override")
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	handleData(w, r, "update")
}

func getCMEBRRNYRR() [5]cmebrrny.ReferenceRate {
	if len(asset_rr) > 0 {
		// Cache the value once every 24 hours
		firstDate := time.Now()
		secondDate := asset_rr[0].Date
		difference := firstDate.Sub(secondDate)
		if difference.Hours() < 24 {
			return asset_rr
		}
	}

	rr, err := cmebrrny.GetBRRYNY()
	if err != nil {
		log.Println("Get Reference Rate error:", err)
		if len(asset_rr) > 0 {
			return asset_rr
		} else {
			return [5]cmebrrny.ReferenceRate{}
		}
	}

	asset_rr = rr

	log.Println("Set Reference Rate:", asset_rr)

	return asset_rr
}

func postDiscord(msg string) {
	blockEmbed := embed{Description: msg}
	embeds := []embed{blockEmbed}
	jsonReq := payload{Username: avatarUsername, AvatarURL: avatarURL, Embeds: embeds}

	jsonStr, _ := json.Marshal(jsonReq)
	log.Println("Discord POST:", string(jsonStr))

	req, _ := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
}

func postTweet(msg string) {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           os.Getenv(OAuthTokenEnvKeyName),
		OAuthTokenSecret:     os.Getenv(OAuthTokenSecretEnvKeyName),
	}

	c, err := gotwi.NewClient(in)
	if err != nil {
		log.Println(err)
		return
	}

	p := &gotwiTypes.CreateInput{
		Text: gotwi.String(msg),
	}

	// Replace newline characters with spaces
	logStr := strings.ReplaceAll(msg, "\n", " ")
	log.Println("X Tweet:", logStr)

	_, err = managetweet.Create(context.Background(), c, p)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
