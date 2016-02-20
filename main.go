package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	notifyUrl       string
	kubelet         bool
	kubeletEndpoint string
	timeout         time.Duration
	urls            []string
)

type notifyData struct {
	Status   string
	Reason   string
	UniqueId string
	Data     string
}

func init() {
	flag.StringVar(&notifyUrl, "notify-url", "", "CloudFormation notify url")
	flag.BoolVar(&kubelet, "kubelet", true, "Check kubelet healthz endpoint")
	flag.StringVar(&kubeletEndpoint, "kubelet-endpoint",
		"https://127.0.0.1:10250/healthz", "Kubelet healthz endpoint url")
	flag.DurationVar(&timeout, "timeout", 5*time.Minute, "Node check total timeout")
}

func main() {
	flag.Parse()
	if notifyUrl == "" {
		log.Fatalln("notifyUrl must be set.")
	}

	if kubelet {
		urls = append(urls, kubeletEndpoint)
	}

	ch := make(chan int)
	// Range over urls and check them
	for _, url := range urls {
		go checkUrl(url, ch)
	}

	count := 0
	defer close(ch)
	for {
		select {
		case i := <-ch:
			// Count successful checks
			count += i
			if len(urls) == count {
				notify(notifyUrl, "SUCCESS")
				os.Exit(0)
			}
		case <-time.After(timeout):
			log.Fatalln("Timeout reached. Exiting..")
		}
	}
}

func notify(url, status string) {
	data := notifyData{
		Status:   status,
		Reason:   "",
		UniqueId: "DOESNOTMATTER",
		Data:     "",
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Sending a %q notification to CloudFormation.\n", status)
	resp, err := http.Post(url, "", bytes.NewBuffer(b))
	if err != nil {
		log.Fatalf("Failed to notify CloudFormation: %q.\n", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to notify CloudFormation: %q.\n", resp.Status)
	}
	return
}

// checkUrl runs in a loop and checks given url until it returns HTTP/200. An
// int 1 is then sent to ch channel.
func checkUrl(url string, ch chan int) {
	for {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Endpoint is unhealthy: %q. %q.\n", url, err)
			time.Sleep(3 * time.Second)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			log.Printf("Endpoint is healthy: %q.\n", url)
			ch <- 1
			return
		} else {
			log.Println(resp.Status)
			time.Sleep(3 * time.Second)
			continue
		}
	}
}
