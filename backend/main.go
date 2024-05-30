package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

type Response struct {
	BucketName, BucketPrefix, Method string
	PresignedUrls                    []string
}

func formatResponse(responses []*v4.PresignedHTTPRequest, bucketName, prefix, method string) Response {
	response := Response{
		BucketName:   bucketName,
		BucketPrefix: prefix,
		Method:       method,
	}
	for _, res := range responses {
		response.PresignedUrls = append(response.PresignedUrls, res.URL)
	}
	return response
}

func GetObjectKeys(client *s3.Client, bucketName string, prefix string) ([]string, error) {
	// Get the first page of results for ListObjectsV2 for a bucket
	result, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		log.Fatal(err)
	}

	// Copy resulting keys into a string
	output := make([]string, 0)
	for _, object := range result.Contents {
		key := aws.ToString(object.Key)
		if key != prefix {
			output = append(output, key)
		}
	}

	return output, err
}

func GetPresignedURL(client *s3.PresignClient, bucketName string, key string, lifetime int) (*v4.PresignedHTTPRequest, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	expiry_callback := s3.WithPresignExpires(time.Duration(lifetime) * time.Second)

	url, err := client.PresignGetObject(context.TODO(), input, expiry_callback)

	return url, err
}

func GetPresignedURLS(client *s3.PresignClient, bucketName string, keys []string, lifetime int) []*v4.PresignedHTTPRequest {
	urls := make([]*v4.PresignedHTTPRequest, len(keys))
	for i, key := range keys {
		url, err := GetPresignedURL(client, bucketName, key, lifetime)
		if err != nil {
			log.Fatal(err)
		}
		urls[i] = url
	}
	return urls
}

func main() {
	// Load environment variables from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Store environment variables
	BUCKET_NAME := os.Getenv("S3_TEST_BUCKET_NAME")
	PREFIX := os.Getenv("S3_TEST_PREFIX")
	LIFETIME_SECS, err := strconv.Atoi(os.Getenv("S3_TEST_LIFETIME"))

	if err != nil {
		LIFETIME_SECS = 3600
	}

	// Load shared AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-1"))
	if err != nil {
		log.Fatal(err)
	}

	// Create an S3 service client
	client := s3.NewFromConfig(cfg)

	keys, err := GetObjectKeys(client, BUCKET_NAME, PREFIX)
	if err != nil {
		log.Fatal(err)
	}

	presignClient := s3.NewPresignClient(client)
	presignedURLS := GetPresignedURLS(presignClient, BUCKET_NAME, keys, LIFETIME_SECS)

	res := formatResponse(presignedURLS, BUCKET_NAME, PREFIX, "GET")
	json_response, _ := json.Marshal(&res)
	fmt.Printf("json_response: %s\n", json_response)
}
