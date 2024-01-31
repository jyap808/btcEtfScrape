package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jyap808/btcEtfScrape/gbtc"
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

	// track the current Result
	result gbtc.Result
)

func init() {
	flag.StringVar(&webhookURL, "webhookURL", "https://discord.com/api/webhooks/", "Webhook URL")
	flag.StringVar(&avatarUsername, "avatarUsername", "Annalee Call", "Avatar username")
	flag.StringVar(&avatarURL, "avatarURL", "https://static1.personality-database.com/profile_images/6604632de9954b4d99575e56404bd8b7.png", "Avatar image URL")
	flag.Parse()
}

func main() {
	for {
		newResult := gbtc.Collect()

		if newResult.TotalBitcoinInTrust != result.TotalBitcoinInTrust {
			if result.TotalBitcoinInTrust == 0 {
				// initialize
				result = newResult
				log.Println("Initialize:", result)
			} else {
				// compare
				bitcoinDiff := result.TotalBitcoinInTrust - newResult.TotalBitcoinInTrust
				aumDiff := result.Aum - newResult.Aum
				layout := "01/02/2006"
				formattedTime := result.Date.Format(layout)
				msg := fmt.Sprintf("GBTC %s\nChange - Bitcoin: %.1f, AUM: $%.1f\nTotal  - Bitcoin: %.1f, AUM: $%.1f", formattedTime, bitcoinDiff, aumDiff, result.TotalBitcoinInTrust, result.Aum)

				postEvent(msg)

				result = newResult
			}
		}

		time.Sleep(time.Minute * 5)
	}
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
