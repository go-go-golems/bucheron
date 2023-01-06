package pkg

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"io"
	"os"
	"path/filepath"
	"time"
)

type UploadSettings struct {
	// The AWS region that the bucket is in.
	Region      string
	Bucket      string
	Credentials *Credentials
}

type UploadData struct {
	Files    []string
	Comment  string
	Metadata map[string]interface{}
}

type UploadProgress struct {
	StepProgress float64
	Step         string
}

type ReaderWithProgress struct {
	reader     io.ReadSeeker
	progressCh chan UploadProgress
	stepName   string
	totalBytes int64
	readBytes  int64
}

func (r *ReaderWithProgress) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.readBytes = offset
	case io.SeekCurrent:
		r.readBytes += offset
	case io.SeekEnd:
		r.readBytes = r.totalBytes + offset

	}
	return r.reader.Seek(offset, whence)
}

func NewReaderWithProgress(
	reader io.ReadSeeker,
	progressCh chan UploadProgress,
	totalBytes int64,
	stepName string,
) *ReaderWithProgress {
	return &ReaderWithProgress{
		reader:     reader,
		progressCh: progressCh,
		totalBytes: totalBytes,
		readBytes:  0,
		stepName:   stepName,
	}
}

func (r ReaderWithProgress) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if err != nil {
		return
	}
	r.readBytes += int64(n)
	r.progressCh <- UploadProgress{
		StepProgress: float64(r.readBytes) / float64(r.totalBytes),
		Step:         r.stepName,
	}
	return
}

func UploadLogs(ctx context.Context, settings *UploadSettings, data *UploadData, progressCh chan UploadProgress) error {
	// Set up a new AWS session using the temporary AWS credentials.
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			settings.Credentials.AccessKeyID,
			settings.Credentials.SecretAccessKey,
			settings.Credentials.SessionToken),
		Region: aws.String(settings.Region),
	})
	if err != nil {
		return err
	}

	// generate unique key: YYYY-MM-DD--HH-MM-SS--UUID
	baseKey := fmt.Sprintf(
		"%s/%s",
		time.Now().Format("2006-01-02/15-04-05"),
		uuid.New().String())

	// Create a new S3 pkg.
	s3Client := s3.New(sess)

	s3Metadata := make(map[string]*string)
	for key, value := range data.Metadata {
		s3Metadata[key] = aws.String(fmt.Sprintf("%v", value))
	}
	s3Metadata["Comment"] = aws.String(data.Comment)

	for i, file := range data.Files {
		err := func() error {
			fp, err := os.Open(file)
			if err != nil {
				return err
			}
			defer fp.Close()
			fileSize, err := fp.Stat()
			if err != nil {
				return err
			}

			baseName := filepath.Base(file)
			// Open the file.
			fileReader := NewReaderWithProgress(
				fp,
				progressCh,
				fileSize.Size(),
				fmt.Sprintf("%d/%d Uploading %s", i+1, len(data.Files), baseName),
			)

			// Upload the file to S3.
			_, err = s3Client.PutObjectWithContext(
				ctx,
				&s3.PutObjectInput{
					Bucket:   aws.String(settings.Bucket),
					Key:      aws.String(baseKey + "/" + baseName),
					Body:     fileReader,
					Metadata: s3Metadata,
				})
			if err != nil {
				return err
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	fmt.Println("Successfully uploaded file to S3")
	return nil
}
