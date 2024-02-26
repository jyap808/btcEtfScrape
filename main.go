package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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

var (
	webhookURL string

	avatarUsername string
	avatarURL      string

	// track
	arkbResult    types.Result
	bitbResult    types.Result
	brrrResult    types.Result
	ezbcResult    types.Result
	gbtcResult    types.Result
	hodlResult    types.Result
	ibitResult    types.Result
	cmebrrnyPrice float64

	// polling intervals
	pollMinutes  int = 5
	backoffHours int = 12

	tickerDescription = map[string]string{
		"ARKB": "Ark/21 Shares",
		"BITB": "Bitwise",
		"BRRR": "Valkyrie",
		"EZBC": "Franklin",
		"GBTC": "Grayscale",
		"HODL": "VanEck",
		"IBIT": "BlackRock",
	}
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
	var wg sync.WaitGroup

	// Increment the WaitGroup counter for each scraping function
	wg.Add(7)

	// Launch goroutines for scraping functions
	go handleFund(&wg, funds.ArkbCollect, arkbResult, "ARKB")
	go handleFund(&wg, funds.BitbCollect, bitbResult, "BITB")
	go handleFund(&wg, funds.BrrrCollect, brrrResult, "BRRR")
	go handleFund(&wg, funds.EzbcCollect, ezbcResult, "EZBC")
	go handleFund(&wg, funds.GbtcCollect, gbtcResult, "GBTC")
	go handleFund(&wg, funds.HodlCollect, hodlResult, "HODL")
	go handleFund(&wg, funds.IbitCollect, ibitResult, "IBIT")

	// Wait for all goroutines to finish
	wg.Wait()

	log.Println("All scraping functions have finished.")
}

// Generic handler
func handleFund(wg *sync.WaitGroup, collector func() types.Result, fundResult types.Result, ticker string) {
	defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes

	for {
		newResult := collector()

		if newResult.TotalBitcoin != fundResult.TotalBitcoin {
			if fundResult.TotalBitcoin == 0 {
				// initialize
				fundResult = newResult
				log.Printf("Initialize %s: %+v", ticker, fundResult)
			} else {
				// compare
				bitcoinDiff := newResult.TotalBitcoin - fundResult.TotalBitcoin
				cmebrrnyPrice = updateCMEBRRNYPrice()
				flowDiff := bitcoinDiff * cmebrrnyPrice

				header := ticker
				if newResult.Date != (time.Time{}) {
					layout := "01/02/2006"
					formattedTime := newResult.Date.Format(layout)

					header = fmt.Sprintf("%s %s", ticker, formattedTime)
				}

				msg := fmt.Sprintf("%s\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS Flow: $%.1f, CMEBRRNY: $%.1f",
					header, bitcoinDiff, newResult.TotalBitcoin,
					flowDiff, cmebrrnyPrice)

				postEvent(msg)

				flowEmoji := "🚀"
				if flowDiff < 0 {
					flowEmoji = "👎"
				}

				xMsg := fmt.Sprintf("%s $%s\n\n%s FLOW: %s BTC, $%s\n🏦 TOTAL Bitcoin in Trust: %s $BTC",
					tickerDescription[ticker], ticker,
					flowEmoji, humanize.CommafWithDigits(bitcoinDiff, 2), humanize.CommafWithDigits(flowDiff, 0),
					humanize.CommafWithDigits(newResult.TotalBitcoin, 1))

				postTweet(xMsg)

				fundResult = newResult

				time.Sleep(time.Hour * time.Duration(backoffHours))
			}
		}

		time.Sleep(time.Minute * time.Duration(pollMinutes))
	}
}

func updateCMEBRRNYPrice() float64 {
	cmebrrnyPrice = cmebrrny.GetBRRYNY().Value
	return cmebrrnyPrice
}

func postEvent(msg string) {
	blockEmbed := embed{Description: msg}
	embeds := []embed{blockEmbed}
	jsonReq := payload{Username: avatarUsername, AvatarURL: avatarURL, Embeds: embeds}

	jsonStr, _ := json.Marshal(jsonReq)
	log.Println("JSON POST:", string(jsonStr))

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

	_, err = managetweet.Create(context.Background(), c, p)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
