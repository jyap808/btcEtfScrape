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

	"github.com/jyap808/btcEtfScrape/arkb"
	"github.com/jyap808/btcEtfScrape/bitb"
	"github.com/jyap808/btcEtfScrape/brrr"
	"github.com/jyap808/btcEtfScrape/cmebrrny"
	"github.com/jyap808/btcEtfScrape/ezbc"
	"github.com/jyap808/btcEtfScrape/gbtc"
	"github.com/jyap808/btcEtfScrape/hodl"
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
	arkbResult    arkb.Result
	bitbResult    bitb.Result
	brrrResult    brrr.Result
	ezbcResult    ezbc.Result
	gbtcResult    gbtc.Result
	hodlResult    hodl.Result
	ibitResult    ibit.Result
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
	wg.Add(7)

	// Launch goroutines for scraping functions
	go handleArkb(&wg)
	go handleBitb(&wg)
	go handleBrrr(&wg)
	go handleEzbc(&wg)
	go handleGbtc(&wg)
	go handleHodl(&wg)
	go handleIbit(&wg)

	// Wait for all goroutines to finish
	wg.Wait()

	log.Println("All scraping functions have finished.")
}

func handleArkb(wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes

	for {
		newResult := arkb.Collect()

		if newResult.TotalBitcoin != arkbResult.TotalBitcoin {
			if arkbResult.TotalBitcoin == 0 {
				// initialize
				arkbResult = newResult
				log.Printf("Initialize ARKB: %+v", arkbResult)
			} else {
				// compare
				bitcoinDiff := newResult.TotalBitcoin - arkbResult.TotalBitcoin
				cmebrrnyPrice = updateCMEBRRNYPrice()
				flowDiff := bitcoinDiff * cmebrrnyPrice
				layout := "01/02/2006"
				formattedTime := newResult.Date.Format(layout)
				msg := fmt.Sprintf("ARKB %s\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS Flow: $%.1f, CMEBRRNY: $%.1f",
					formattedTime, bitcoinDiff, newResult.TotalBitcoin,
					flowDiff, cmebrrnyPrice)

				postEvent(msg)

				arkbResult = newResult

				time.Sleep(time.Hour * time.Duration(backoffHours))
			}
		}

		time.Sleep(time.Minute * time.Duration(pollMinutes))
	}
}

func handleBitb(wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes

	for {
		newResult := bitb.Collect()

		if newResult.TotalBitcoin != bitbResult.TotalBitcoin {
			if bitbResult.TotalBitcoin == 0 {
				// initialize
				bitbResult = newResult
				log.Printf("Initialize BITB: %+v", bitbResult)
			} else {
				// compare
				bitcoinDiff := newResult.TotalBitcoin - bitbResult.TotalBitcoin
				cmebrrnyPrice = updateCMEBRRNYPrice()
				flowDiff := bitcoinDiff * cmebrrnyPrice
				msg := fmt.Sprintf("BITB\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS Flow: $%.1f, CMEBRRNY: $%.1f",
					bitcoinDiff, newResult.TotalBitcoin,
					flowDiff, cmebrrnyPrice)

				postEvent(msg)

				bitbResult = newResult

				time.Sleep(time.Hour * time.Duration(backoffHours))
			}
		}

		time.Sleep(time.Minute * time.Duration(pollMinutes))
	}
}

func handleBrrr(wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes

	for {
		newResult := brrr.Collect()

		if newResult.TotalBitcoin != brrrResult.TotalBitcoin {
			if brrrResult.TotalBitcoin == 0 {
				// initialize
				brrrResult = newResult
				log.Printf("Initialize BRRR: %+v", brrrResult)
			} else {
				// compare
				bitcoinDiff := newResult.TotalBitcoin - brrrResult.TotalBitcoin
				cmebrrnyPrice = updateCMEBRRNYPrice()
				flowDiff := bitcoinDiff * cmebrrnyPrice
				msg := fmt.Sprintf("BRRR\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS Flow: $%.1f, CMEBRRNY: $%.1f",
					bitcoinDiff, newResult.TotalBitcoin,
					flowDiff, cmebrrnyPrice)

				postEvent(msg)

				brrrResult = newResult

				time.Sleep(time.Hour * time.Duration(backoffHours))
			}
		}

		time.Sleep(time.Minute * time.Duration(pollMinutes))
	}
}

func handleEzbc(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		newResult := ezbc.Collect()

		if newResult.TotalBitcoin != ezbcResult.TotalBitcoin {
			if ezbcResult.TotalBitcoin == 0 {
				// initialize
				ezbcResult = newResult
				log.Printf("Initialize EZBC: %+v", ezbcResult)
			} else {
				// compare
				bitcoinDiff := newResult.TotalBitcoin - ezbcResult.TotalBitcoin
				cmebrrnyPrice = updateCMEBRRNYPrice()
				flowDiff := bitcoinDiff * cmebrrnyPrice
				layout := "01/02/2006"
				formattedTime := newResult.Date.Format(layout)
				msg := fmt.Sprintf("EZBC %s\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS Flow: $%.1f, CMEBRRNY: $%.1f",
					formattedTime, bitcoinDiff, newResult.TotalBitcoin,
					flowDiff, cmebrrnyPrice)

				postEvent(msg)

				ezbcResult = newResult

				time.Sleep(time.Hour * time.Duration(backoffHours))
			}
		}

		time.Sleep(time.Minute * time.Duration(pollMinutes))
	}
}

func handleGbtc(wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes

	for {
		newResult := gbtc.Collect()

		if newResult.TotalBitcoin != gbtcResult.TotalBitcoin {
			if gbtcResult.TotalBitcoin == 0 {
				// initialize
				gbtcResult = newResult
				log.Printf("Initialize GBTC: %+v", gbtcResult)
			} else {
				// compare
				bitcoinDiff := newResult.TotalBitcoin - gbtcResult.TotalBitcoin
				cmebrrnyPrice = updateCMEBRRNYPrice()
				flowDiff := bitcoinDiff * cmebrrnyPrice
				layout := "01/02/2006"
				formattedTime := newResult.Date.Format(layout)
				msg := fmt.Sprintf("GBTC %s\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS Flow: $%.1f, CMEBRRNY: $%.1f",
					formattedTime, bitcoinDiff, newResult.TotalBitcoin,
					flowDiff, cmebrrnyPrice)

				postEvent(msg)

				gbtcResult = newResult

				time.Sleep(time.Hour * time.Duration(backoffHours))
			}
		}

		time.Sleep(time.Minute * time.Duration(pollMinutes))
	}
}

func handleHodl(wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes

	for {
		newResult := hodl.Collect()

		if newResult.TotalBitcoin != hodlResult.TotalBitcoin {
			if hodlResult.TotalBitcoin == 0 {
				// initialize
				hodlResult = newResult
				log.Printf("Initialize HODL: %+v", hodlResult)
			} else {
				// compare
				bitcoinDiff := newResult.TotalBitcoin - hodlResult.TotalBitcoin
				cmebrrnyPrice = updateCMEBRRNYPrice()
				flowDiff := bitcoinDiff * cmebrrnyPrice
				layout := "01/02/2006"
				formattedTime := newResult.Date.Format(layout)
				msg := fmt.Sprintf("HODL %s\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS Flow: $%.1f, CMEBRRNY: $%.1f",
					formattedTime, bitcoinDiff, newResult.TotalBitcoin,
					flowDiff, cmebrrnyPrice)

				postEvent(msg)

				hodlResult = newResult

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

		if newResult.TotalBitcoin != ibitResult.TotalBitcoin {
			if ibitResult.TotalBitcoin == 0 {
				// initialize
				ibitResult = newResult
				log.Printf("Initialize IBIT: %+v", ibitResult)
			} else {
				// compare
				bitcoinDiff := newResult.TotalBitcoin - ibitResult.TotalBitcoin
				cmebrrnyPrice = updateCMEBRRNYPrice()
				flowDiff := bitcoinDiff * cmebrrnyPrice
				msg := fmt.Sprintf("IBIT\nCHANGE Bitcoin: %.1f\nTOTAL Bitcoin: %.1f\nDETAILS Flow: $%.1f, CMEBRRNY: $%.1f",
					bitcoinDiff, newResult.TotalBitcoin,
					flowDiff, cmebrrnyPrice)

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
