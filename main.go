package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3/s3manager"
)

func main() {
	// we expect exactly one CLI argument: the bucket name
	// if that is not provided, exit
	if len(os.Args) != 2 {
		log.Fatalf("Can't continue, no bucket name was provided!")
	}
	bucket := os.Args[1]
	now := time.Now()
	key := fmt.Sprintf("s3echoer-%v", now.Unix())

	userinput, err := userInput()
	if err != nil {
		log.Fatalf("Can't read from stdin: %v", err)
	}
	fmt.Printf("Uploading user input to S3 using %v/%v\n\n", bucket, key)
	err = uploadToS3(bucket, key, userinput)
	if err != nil {
		log.Fatalf("Can't upload to S3: %v", err)
	}
}

// userInput reads from stdin until it sees a CTRL+D.
func userInput() (string, error) {
	rawinput, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(rawinput), nil
}

// uploadToS3 puts the payload into the S3 bucket using the key provided.
func uploadToS3(bucket, key, payload string) error {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(cfg)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(payload),
	})
	return err
}
