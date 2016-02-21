package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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

// Get the resource tag value given the resource id
func getResourceTagValue(id, tag string) string {
	c := ec2.New(session.New(&aws.Config{Region: &region}))
	params := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("resource-id"),
				Values: []*string{
					aws.String(id),
				},
			},
			{
				Name: aws.String("key"),
				Values: []*string{
					aws.String(tag),
				},
			},
		},
	}
	resp, err := c.DescribeTags(params)
	if err != nil {
		log.Printf("Cannot get tag %q of %q resource: %q.\n", tag, id, err)
		return ""
	}
	if len(resp.Tags) > 0 {
		for _, t := range resp.Tags {
			return *t.Value
		}
	}
	log.Printf("Cannot get tag %q of %q resource.\n", tag, id)
	return ""
}
