package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gregdel/pushover"
)

const cloudflareURL = "https://api.cloudflare.com/client/v4/ips"

type cloudflareIPsResult struct {
	IPv4CIDRs []string `json:"ipv4_cidrs"`
	IPv6CIDRs []string `json:"ipv6_cidrs"`
	Etag      string   `json:"etag"`
}

type cloudflareIPs struct {
	Result   cloudflareIPsResult `json:"result"`
	Success  bool                `json:"success"`
	Errors   []string            `json:"errors"`
	Messages []string            `json:"messages"`
}

func main() {
	appKey := os.Getenv("PUSHOVER_APP_KEY")
	groupKey := os.Getenv("PUSHOVER_GROUP_KEY")
	stateFile := os.Getenv("STATE_FILE")

	// Create a new pushover app with a token
	app := pushover.New(appKey)

	// Create a new recipient
	recipient := pushover.NewRecipient(groupKey)

	// Create the change notification message
	changeMessage := &pushover.Message{
		Message:  "The Cloudflare API is advertising a CIDR block change.",
		Title:    "Cloudflare CIDR Blocks Changed",
		Priority: pushover.PriorityEmergency,
		URL:      "https://www.cloudflare.com/ips/",
		Retry:    5 * time.Minute,
		Expire:   24 * time.Hour,
	}

	// Create a service error message to send
	errMessage := &pushover.Message{
		Message:  "The process watching for Cloudflare CIDR block changes failed.",
		Title:    "Cloudflare CIDR Watch Failed",
		Priority: pushover.PriorityNormal,
		URL:      "https://www.cloudflare.com/ips/",
	}

	cloudflareClient := http.Client{
		Timeout: time.Second * 30, // Timeout after 30 seconds
	}

	// Read last seen etag
	etag, _ := ioutil.ReadFile(stateFile)
	log.Println(fmt.Sprintf("Initial etag: %s", string(etag)))

	for {
		log.Println("Checking Cloudflare APIs")
		req, err := http.NewRequest(http.MethodGet, cloudflareURL, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("User-Agent", "cfcidrwatch")

		res, err := cloudflareClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}

		ips := cloudflareIPs{}
		err = json.Unmarshal(body, &ips)
		if err != nil {
			log.Fatal(err)
		}

		if etag == nil {
			log.Println(fmt.Sprintf("Saw first etag: %s", ips.Result.Etag))
			// Save the etag to disk
			err = ioutil.WriteFile(stateFile, []byte(ips.Result.Etag), 0644)
			if err != nil {
				log.Fatal(err)
			}
		} else if !ips.Success {
			log.Println(fmt.Sprintf("Failed to call Cloudflare API: %v", ips.Errors))
			_, err := app.SendMessage(errMessage, recipient)
			if err != nil {
				log.Fatal(err)
			}
		} else if ips.Result.Etag != string(etag) {
			log.Println(fmt.Sprintf("Saw new etag: %s", ips.Result.Etag))
			// Send the message to the recipient
			_, err := app.SendMessage(changeMessage, recipient)
			if err != nil {
				log.Fatal(err)
			}

			// Save the latest etag to disk
			err = ioutil.WriteFile(stateFile, []byte(ips.Result.Etag), 0644)
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Println("Completed API check")
		time.Sleep(time.Hour)
	}
}
