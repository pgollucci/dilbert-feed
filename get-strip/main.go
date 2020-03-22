package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kelseyhightower/envconfig"

	"github.com/mlafeldt/dilbert-feed/dilbert"
)

// Input is the input passed to the Lambda function.
type Input struct {
	Date string `json:"date"`
}

// Output is the output returned by the Lambda function.
type Output struct {
	*dilbert.Comic
	UploadURL string `json:"upload_url"`
}

func main() {
	lambda.Start(handler)
}

func handler(input Input) (*Output, error) {
	var env struct {
		BucketName string `envconfig:"BUCKET_NAME" required:"true"`
		StripsDir  string `envconfig:"STRIPS_DIR" required:"true"`
	}
	if err := envconfig.Process("", &env); err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] env = %+v", env)

	var date string
	if input.Date != "" {
		date = strings.TrimSpace(input.Date)
		if len(date) != 10 {
			return nil, fmt.Errorf("input date %q has invalid length", date)
		}
		if len(strings.Split(date, "-")) != 3 {
			return nil, fmt.Errorf("input date %q has invalid format", date)
		}
	}

	comic, err := dilbert.NewComic(date)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] comic = %+v", comic)
	log.Printf("[INFO] Uploading strip %s to bucket %q ...", comic.StripURL, env.BucketName)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(comic.ImageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	stripPath := fmt.Sprintf("%s%s.gif", env.StripsDir, comic.Date)
	stripURL, err := uploadStrip(resp.Body, env.BucketName, stripPath, comic.Title)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Upload completed: %s", stripURL)
	return &Output{comic, stripURL}, nil
}

func uploadStrip(r io.Reader, bucketName, stripPath, title string) (string, error) {
	sess, err := session.NewSession()
	if err != nil {
		return "", err
	}

	upload, err := s3manager.NewUploader(sess).Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(stripPath),
		ContentType: aws.String("image/gif"),
		// Add strip title to metadata for gen-feed to create nicer RSS feed entries
		Metadata: map[string]*string{
			"Title": aws.String(title),
		},
		Body: r,
	})
	if err != nil {
		return "", err
	}

	return upload.Location, nil
}
