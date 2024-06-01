package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	EnvBucket   = "S3_TEST_BUCKET_NAME"
	EnvPrefix   = "S3_TEST_PREFIX"
	EnvLifetime = "S3_TEST_LIFETIME"
)

type Response struct {
	BucketName    string   `json:"bucket_name"`
	BucketPrefix  string   `json:"bucket_prefix"`
	Method        string   `json:"method"`
	PresignedUrls []string `json:"presigned_urls"`
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

func GetObjectKeys(client *s3.Client, bucketName, prefix string) ([]string, error) {
	result, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in bucket %s with prefix %s: %w", bucketName, prefix, err)
	}

	var keys []string
	for _, object := range result.Contents {
		key := aws.ToString(object.Key)
		if key != prefix {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

func GetPresignedURL(client *s3.PresignClient, bucketName, key string, lifetime int) (*v4.PresignedHTTPRequest, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	url, err := client.PresignGetObject(context.TODO(), input, s3.WithPresignExpires(time.Duration(lifetime)*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL for key %s: %w", key, err)
	}
	return url, nil
}

func GetPresignedURLS(client *s3.PresignClient, bucketName string, keys []string, lifetime int) ([]*v4.PresignedHTTPRequest, error) {
	urls := make([]*v4.PresignedHTTPRequest, len(keys))
	for i, key := range keys {
		url, err := GetPresignedURL(client, bucketName, key, lifetime)
		if err != nil {
			return nil, err
		}
		urls[i] = url
	}
	return urls, nil
}

func loadEnvironmentVariables() (string, string, int, error) {
	// err := godotenv.Load()
	// if err != nil {
	// 	return "", "", 0, fmt.Errorf("error loading .env file: %w", err)
	// }

	bucketName := os.Getenv(EnvBucket)
	if bucketName == "" {
		return "", "", 0, fmt.Errorf("environment variable %s is required", EnvBucket)
	}

	prefix := os.Getenv(EnvPrefix)
	if prefix == "" {
		return "", "", 0, fmt.Errorf("environment variable %s is required", EnvPrefix)
	}

	lifetimeSecs, err := strconv.Atoi(os.Getenv(EnvLifetime))
	if err != nil {
		log.Printf("invalid %s value, defaulting to 3600 seconds", EnvLifetime)
		lifetimeSecs = 3600
	}

	return bucketName, prefix, lifetimeSecs, nil
}

func loadConfig(region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to load SDK config, %w", err)
	}
	return cfg, nil
}

func GetPresignedTestImages() (Response, error) {
	bucketName, prefix, lifetimeSecs, err := loadEnvironmentVariables()
	if err != nil {
		return Response{}, fmt.Errorf("could not load environment variables: %w", err)
	}

	cfg, err := loadConfig("us-west-1")
	if err != nil {
		return Response{}, fmt.Errorf("could not load AWS configuration: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	keys, err := GetObjectKeys(client, bucketName, prefix)
	if err != nil {
		return Response{}, fmt.Errorf("could not get object keys: %w", err)
	}

	presignClient := s3.NewPresignClient(client)
	presignedURLs, err := GetPresignedURLS(presignClient, bucketName, keys, lifetimeSecs)
	if err != nil {
		return Response{}, fmt.Errorf("could not get presigned URLs: %w", err)
	}

	return formatResponse(presignedURLs, bucketName, prefix, "GET"), nil
}

func HandleRequest(ctx context.Context, event interface{}) (string, error) {
	response, err := GetPresignedTestImages()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatalf("error marshalling response to JSON: %v", err)
	}

	return string(jsonResponse), nil
}

func main() {
	lambda.Start(HandleRequest)
}
