package main

import (
	"crypto/tls"
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	stack                 string
	resource              string
	region                string
	uniqueID              string
	kubelet               bool
	kubeletEndpoint       string
	timeout               time.Duration
	urls                  []string
	insecureSkipTLSVerify bool
)

func init() {
	flag.StringVar(&stack, "stack", "", "CloudFormation stack name.")
	flag.StringVar(&resource, "resource", "", "CloudFormation resource name.")
	flag.StringVar(&region, "region", "", "AWS region")
	flag.StringVar(&uniqueID, "unique-id", "",
		"A unique ID. When you signal EC2 instances or Auto Scaling groups, specify the instance ID.")
	flag.BoolVar(&kubelet, "kubelet", true, "Check kubelet healthz endpoint")
	flag.StringVar(&kubeletEndpoint, "kubelet-endpoint",
		"https://127.0.0.1:10250/healthz", "Kubelet healthz endpoint url")
	flag.DurationVar(&timeout, "timeout", 5*time.Minute, "Node check total timeout")
	flag.BoolVar(&insecureSkipTLSVerify, "insecure-skip-tls-verify", false,
		"If true, endpoint's certificate will not be checked for validity.")
}

func main() {
	flag.Parse()
	if stack == "" {
		log.Fatalln("Stack name must be specified.")
	}
	if resource == "" {
		log.Fatalln("Logical resource name must be specified.")
	}
	if region == "" {
		if isMetadataAvailable() {
			region = getRegion()
		} else {
			log.Fatalln("EC2 metadata service is not available. Specify region.")
		}
	}
	if uniqueID == "" {
		if isMetadataAvailable() {
			uniqueID = getInstanceID()
		} else {
			log.Fatalln("EC2 metadata service is not available. Specify unique ID.")
		}
	}

	if kubelet {
		urls = append(urls, kubeletEndpoint)
	}

	ch := make(chan int)
	// Range over urls and check them
	for _, u := range urls {
		go checkURL(u, ch)
	}

	count := 0
	defer close(ch)
	for {
		select {
		case i := <-ch:
			// Count successful checks
			count += i
			if len(urls) == count {
				sendSignal("SUCCESS")
				os.Exit(0)
			}
		case <-time.After(timeout):
			log.Fatalln("Timeout reached. Exiting..")
		}
	}
}

func sendSignal(status string) {
	cf := cloudformation.New(session.New(&aws.Config{Region: &region}))
	params := &cloudformation.SignalResourceInput{
		LogicalResourceId: &resource,
		StackName:         &stack,
		Status:            &status,
		UniqueId:          &uniqueID,
	}
	_, err := cf.SignalResource(params)
	if err != nil {
		log.Fatalf("Failed to signal CloudFormation: %q.\n", err.Error())
	}
	log.Printf("Sent a %q signal to CloudFormation.\n", status)
	return
}

// checkUrl runs in a loop and checks given url until it returns HTTP/200. An
// int 1 is then sent to ch channel.
func checkURL(url string, ch chan int) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipTLSVerify},
	}
	client := &http.Client{Transport: tr}
	for {
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("Endpoint is unhealthy: %q. %q.\n", url, err)
			time.Sleep(3 * time.Second)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			log.Printf("Endpoint is healthy: %q.\n", url)
			ch <- 1
			return
		}
		log.Println(resp.Status)
		time.Sleep(3 * time.Second)
		continue
	}
}
