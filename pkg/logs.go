package pkg

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"time"
)

type LogEntry struct {
	Key      string
	Date     time.Time
	Size     int64
	Comment  string
	Metadata map[string]string
	FileName string
	UUID     uuid.UUID
}

type ListFilter struct {
	From   *time.Time
	To     *time.Time
	Globs  []string
	Prefix string
}

func DownloadLogs(
	ctx context.Context,
	settings *BucketSettings,
	filter *ListFilter,
	outputDirectory string,
	progressCh chan ProgressEvent,
) error {
	defer close(progressCh)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(settings.Region),
	})
	if err != nil {
		return err
	}

	downloader := s3manager.NewDownloader(sess)
	entryCh2 := make(chan LogEntry)

	errGroup, ctx2 := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		return ListLogBucketKeys(ctx2, settings, filter, entryCh2)
	})

	errGroup.Go(func() error {
		for {
			select {
			case entry, ok := <-entryCh2:
				if !ok {
					return nil
				}

				fileName := filepath.Join(outputDirectory, entry.Key)
				// ensure directory exists
				err := os.MkdirAll(filepath.Dir(fileName), 0755)
				if err != nil {
					return err
				}

				f, err := os.Create(fileName)
				if err != nil {
					return err
				}

				wf := NewWriterWithProgress(f, progressCh, entry.Size, fmt.Sprintf("Downloading %s", entry.Key))

				// TODO(manuel, 2023-01-06) Could be parallelized
				// https://github.com/go-go-golems/bucheron/issues/1
				log.Debug().Str("key", entry.Key).Str("fileName", fileName).Msg("Downloading")
				_, err = downloader.DownloadWithContext(ctx, wf, &s3.GetObjectInput{
					Bucket: aws.String(settings.Bucket),
					Key:    aws.String(entry.Key),
				})
				if err != nil {
					return err
				}

			case <-ctx2.Done():
				return ctx2.Err()
			}
		}
	})

	return errGroup.Wait()
}

func ListLogBucketKeys(
	ctx context.Context,
	settings *BucketSettings,
	filter *ListFilter,
	entryCh chan LogEntry,
) error {
	defer close(entryCh)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(settings.Region),
	})
	if err != nil {
		return err
	}

	s3Client := s3.New(sess)

	var startAfter *string = aws.String("")

	if filter.From != nil {
		startAfter = aws.String(filter.From.Format("2006-01-02"))
		log.Debug().Msgf("Filtering logs from %s", *startAfter)
	}

	var nextContinuationToken *string

	log.Debug().Msgf("Listing keys on bucket %s (region %s)", settings.Bucket, settings.Region)

	for {
		log.Debug().Msgf("Listing objects from %s with continuation token %v",
			*startAfter, nextContinuationToken)

		resp, err := s3Client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(settings.Bucket),
			Prefix: aws.String(filter.Prefix),
			//StartAfter:        startAfter,
			ContinuationToken: nextContinuationToken,
		})

		if err != nil {
			return err
		}

		log.Debug().Msgf("Found %d objects", len(resp.Contents))

		// AppendMatch the objects to the list
		for _, obj := range resp.Contents {
			fileName := filepath.Base(*obj.Key)

			if len(filter.Globs) > 0 {
				log.Debug().Msgf("Checking if %s matches %v", *obj.Key, filter.Globs)
				if !matchAnyGlob(fileName, filter.Globs) {
					continue
				}
			}

			head, err := s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
				Bucket: aws.String(settings.Bucket),
				Key:    obj.Key,
			})
			if err != nil {
				log.Error().Err(err).Msgf("Failed to get metadata for %s", *obj.Key)
				continue
			}

			// key is in format 2006-01-06/15-04-05/UUID/filename
			if len(*obj.Key) < 56 {
				log.Error().Msgf("Invalid key %s", *obj.Key)
				continue
			}
			date := (*obj.Key)[0:19]
			uuidString := (*obj.Key)[20:56]
			parsedUUID, err := uuid.Parse(uuidString)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to parse UUID %s", uuidString)
				continue
			}
			t, err := time.Parse("2006-01-02/15-04-05", date)
			if err != nil {
				log.Error().Msgf("Failed to parse date from key %s", *obj.Key)
				continue
			}

			comment := ""
			metadata := make(map[string]string)
			for k, v := range head.Metadata {
				if k != "Comment" && v != nil {
					metadata[k] = *v
				}
				if k == "Comment" && v != nil {
					comment = *v
				}
			}

			// check if head.LastModified is after endBefore
			if filter.To != nil && head.LastModified.After(*filter.To) {
				log.Debug().Msgf("Skipping %s because it is after %s", *obj.Key, filter.To.Format(time.RFC3339))
				continue
			}

			entryCh <- LogEntry{
				Key:      *obj.Key,
				FileName: fileName,
				Comment:  comment,
				Metadata: metadata,
				Date:     t,
				Size:     *obj.Size,
				UUID:     parsedUUID,
			}
		}

		// Set the continuation token for the next iteration
		contToken := ""
		if resp.NextContinuationToken != nil {
			contToken = *resp.NextContinuationToken
		}

		log.Debug().
			Str("nextContinuationToken", contToken).
			Msg("Setting next continuation token")
		nextContinuationToken = resp.NextContinuationToken

		// Break the loop if there are no more objects to retrieve
		if nextContinuationToken == nil {
			break
		}
	}

	log.Debug().
		Str("bucket", settings.Bucket).
		Str("region", settings.Region).
		Msg("Finished listing keys on bucket")

	return nil
}

func matchAnyGlob(s string, globs []string) bool {
	for _, g := range globs {
		matched, err := filepath.Match(g, s)
		log.Debug().Msgf("Checking if %s matches %s: %v", s, g, matched)
		if matched && err == nil {
			return true
		}
	}

	return false
}
