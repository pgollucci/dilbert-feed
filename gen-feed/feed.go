package main

import (
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/feeds"
	"github.com/mlafeldt/dilbert-feed/dilbert"
)

func generateFeed(w io.Writer, startDate time.Time, feedLength int, tableName string) error {
	feed := &feeds.Feed{
		Title:       "Dilbert",
		Link:        &feeds.Link{Href: "http://dilbert.com"},
		Description: "Dilbert Daily Strip",
	}

	sess, err := session.NewSession()
	if err != nil {
		return err
	}
	svc := dynamodb.New(sess)

	for i := 0; i < feedLength; i++ {
		day := startDate.AddDate(0, 0, -i).Truncate(24 * time.Hour)
		date := fmt.Sprintf("%d-%02d-%02d", day.Year(), day.Month(), day.Day())

		item, err := svc.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"date": {S: aws.String(date)},
			},
		})
		if err != nil {
			return err
		}

		var comic dilbert.Comic

		if err := dynamodbattribute.UnmarshalMap(item.Item, &comic); err != nil {
			return err
		}

		feed.Add(&feeds.Item{
			Title:       comic.Title,
			Link:        &feeds.Link{Href: comic.ImageURL},
			Description: fmt.Sprintf(`<img src="%s">`, comic.ImageURL),
			Id:          comic.ImageURL,
			Created:     day,
		})
	}

	return feed.WriteRss(w)
}

func uploadFeed(r io.Reader, bucketName, feedPath string) (string, error) {
	sess, err := session.NewSession()
	if err != nil {
		return "", err
	}

	upload, err := s3manager.NewUploader(sess).Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(feedPath),
		Body:        r,
		ContentType: aws.String("text/xml; charset=utf-8"),
	})
	if err != nil {
		return "", err
	}

	return upload.Location, nil
}
