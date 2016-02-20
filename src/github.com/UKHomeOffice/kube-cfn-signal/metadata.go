package main

import (
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
)

func getRegion() string {
	c := ec2metadata.New(session.New())
	r, err := c.Region()
	if err != nil {
		log.Fatalln("Failed to get region from the metadata service.")
		return ""
	}
	return r
}

func getInstanceID() string {
	c := ec2metadata.New(session.New())
	i, err := c.GetMetadata("instance-id")
	if err != nil {
		log.Fatalln("Failed to get instance ID from the metadata service.")
		return ""
	}
	return i
}

func isMetadataAvailable() bool {
	c := ec2metadata.New(session.New())
	return c.Available()
}
