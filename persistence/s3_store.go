package persistence

import (
	"bytes"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Store struct {
	sess       *session.Session
	bucketName string
	key        string
}

func NewS3Store(
	region string,
	bucketName string, key string,
) (PersistentStore, error) {
	sess, err := session.NewSession(&aws.Config{Region: &region})
	if err != nil {
		return nil, fmt.Errorf("create s3 store: %w", err)
	}

	return &s3Store{
		sess:       sess,
		bucketName: bucketName,
		key:        key,
	}, nil
}

func (s *s3Store) Load() ([]byte, error) {
	d := s3manager.NewDownloader(s.sess)
	buf := aws.NewWriteAtBuffer([]byte{})
	_, err := d.Download(buf, &s3.GetObjectInput{Bucket: &s.bucketName, Key: &s.key})
	if err != nil {
		return nil, fmt.Errorf("download object: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *s3Store) ModTime() (time.Time, bool, error) {
	client := s3.New(s.sess)
	obj, err := client.GetObject(&s3.GetObjectInput{Bucket: &s.bucketName, Key: &s.key})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchKey:
				return time.Time{}, false, nil
			}
		}
		return time.Time{}, true, fmt.Errorf("get object: %w", err)
	}
	return *obj.LastModified, true, nil
}

func (s *s3Store) Save(data []byte) error {
	u := s3manager.NewUploader(s.sess)
	reader := bytes.NewReader(data)
	_, err := u.Upload(&s3manager.UploadInput{Bucket: &s.bucketName, Key: &s.key, Body: reader})
	if err != nil {
		return fmt.Errorf("upload object: %w", err)
	}
	return nil
}
