package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jyap808/btcEtfScrape/cmebrrny"
	"github.com/jyap808/btcEtfScrape/gbtc"
	"github.com/jyap808/btcEtfScrape/ibit"
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
	gbtcResult    gbtc.Result
	ibitResult    float64
	cmebrrnyPrice float64

	// polling intervals
	pollMinutes  int = 5
	backoffHours int = 12
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
	wg.Add(2)

	// Launch goroutines for scraping functions
	go handleGbtc(&wg)
	go handleIbit(&wg)

	// Wait for all goroutines to finish
	wg.Wait()

	log.Println("All scraping functions have finished.")
}

func handleGbtc(wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes

	for {
		newResult := gbtc.Collect()

		if newResult.TotalBitcoinInTrust != gbtcResult.TotalBitcoinInTrust {
			if gbtcResult.TotalBitcoinInTrust == 0 {
				// initialize
				gbtcResult = newResult
				log.Printf("Initialize GBTC: %+v", gbtcResult)
			} else {
				// compare
				bitcoinDiff := newResult.TotalBitcoinInTrust - gbtcResult.TotalBitcoinInTrust
				cmebrrnyPrice = updateCMEBRRNYPrice()
				flowDiff := bitcoinDiff * cmebrrnyPrice
				layout := "01/02/2006"
				formattedTime := newResult.Date.Format(layout)
				msg := fmt.Sprintf("GBTC %s\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS Flow: $%.1f, CMEBRRNY: $%.1f",
					formattedTime, bitcoinDiff, newResult.TotalBitcoinInTrust,
					flowDiff, cmebrrnyPrice)

				postEvent(msg)

				gbtcResult = newResult

				time.Sleep(time.Hour * time.Duration(backoffHours))
			}
		}

		time.Sleep(time.Minute * time.Duration(pollMinutes))
	}
}

func handleIbit(wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes

	for {
		newResult := ibit.Collect()

		if newResult != ibitResult {
			if ibitResult == 0 {
				// initialize
				ibitResult = newResult
				log.Printf("Initialize IBIT: %+v", ibitResult)
			} else {
				// compare
				bitcoinDiff := newResult - ibitResult
				cmebrrnyPrice = updateCMEBRRNYPrice()
				msg := fmt.Sprintf("IBIT\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS CMEBRRNY: $%.1f",
					bitcoinDiff, newResult, cmebrrnyPrice)

				postEvent(msg)

				ibitResult = newResult

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
