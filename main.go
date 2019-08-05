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
	awsv1 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	s3managerv1 "github.com/aws/aws-sdk-go/service/s3/s3manager"
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

	irp := false
	if irpenv := os.Getenv("ENABLE_IRP"); irpenv != "" {
		irp = true
	}
	switch irp {
	case true:
		err = uploadToS3IRP(bucket, key, userinput)
		if err != nil {
			log.Fatalf("Can't upload to S3: %v", err)
		}
	case false:
		err = uploadToS3(bucket, key, userinput)
		if err != nil {
			log.Fatalf("Can't upload to S3: %v", err)
		}
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

// uploadToS3IRP puts the payload into the S3 bucket using the key provided.
// it uses https://github.com/aws/aws-sdk-go/releases/tag/v1.21.9 with
// https://github.com/aws/aws-sdk-go/pull/2667
func uploadToS3IRP(bucket, key, payload string) error {
	region := "us-west-2"
	if regionenv := os.Getenv("AWS_DEFAULT_REGION"); regionenv != "" {
		region = regionenv
	}
	sess := session.Must(session.NewSession(&awsv1.Config{
		Region: aws.String(region),
	}))
	uploader := s3managerv1.NewUploader(sess)
	_, err := uploader.Upload(&s3managerv1.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(payload),
	})
	return err
}
