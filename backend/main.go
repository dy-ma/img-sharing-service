package main

import (
	"context"
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

const (
	EnvBucket   = "S3_TEST_BUCKET_NAME"
	EnvPrefix   = "S3_TEST_PREFIX"
	EnvLifetime = "S3_Test_LIFETIME"
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
		return nil, err
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

func loadEnvironmentVariables() (string, string, int, error) {
	// Load environment variables from .env
	err := godotenv.Load()
	if err != nil {
		return "", "", 0, err
	}

	// Store environment variables
	bucketName := os.Getenv(EnvBucket)
	prefix := os.Getenv(EnvPrefix)
	lifetimeSecs, err := strconv.Atoi(os.Getenv(EnvLifetime))

	if err != nil {
		lifetimeSecs = 3600
	}

	return bucketName, prefix, lifetimeSecs, nil
}

func loadConfig(region string) (aws.Config, error) {
	return config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
}

func GetPresignedTestImages() (Response, error) {

	BUCKET_NAME, PREFIX, LIFETIME_SECS, err := loadEnvironmentVariables()
	if err != nil {
		return Response{}, fmt.Errorf("couldn't locate environment variables at: %s\n %s\n %s\n ", EnvBucket, EnvPrefix, EnvLifetime)
	}

	// Load shared AWS configuration
	cfg, err := loadConfig("us-west-1")
	if err != nil {
		return Response{}, err
	}

	// Create an S3 service client
	client := s3.NewFromConfig(cfg)

	keys, err := GetObjectKeys(client, BUCKET_NAME, PREFIX)
	if err != nil {
		return Response{}, err
	}

	presignClient := s3.NewPresignClient(client)
	presignedURLS := GetPresignedURLS(presignClient, BUCKET_NAME, keys, LIFETIME_SECS)

	res := formatResponse(presignedURLS, BUCKET_NAME, PREFIX, "GET")

	return res, nil
}

func main() {
	response, err := GetPresignedTestImages()
	if err != nil {
		log.Fatal(err)
		// Return error response
	}

	fmt.Printf("%+v\n", response)
}
