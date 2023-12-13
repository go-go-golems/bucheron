package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/bucheron/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload one or more files to S3",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := viper.GetString("bucket")
		region := viper.GetString("region")
		comment, _ := cmd.Flags().GetString("comment")

		// get temporary credentials over HTTP
		api := viper.GetString("api")
		ctx, cancel := context.WithCancel(context.Background())

		credentials, err := pkg.GetUploadCredentials(ctx, api)
		cobra.CheckErr(err)

		settings := &pkg.BucketSettings{
			Region:      region,
			Bucket:      bucket,
			Credentials: credentials,
		}

		data := &pkg.UploadData{
			Files:    args,
			Comment:  comment,
			Metadata: nil,
		}

		progressCh := make(chan pkg.ProgressEvent)

		ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
		defer stop()

		errGroup, ctx2 := errgroup.WithContext(ctx)
		// Set up a signal handler to cancel the context when the user
		// presses Ctrl+C

		errGroup.Go(func() error {
			defer cancel()
			return pkg.UploadLogs(ctx2, settings, data, progressCh)
		})
		errGroup.Go(func() error {
			for {
				select {
				case progress, ok := <-progressCh:
					if !ok {
						cancel()
						return nil
					}
					fmt.Printf("%s: %f\n", progress.Step, progress.StepProgress)

				case <-ctx2.Done():
					return ctx2.Err()
				}
			}
		})

		err = errGroup.Wait()
		if err != context.Canceled {
			cobra.CheckErr(err)
		}
	},
}

func init() {
	uploadCmd.Flags().StringP("comment", "c", "", "Comment to add to the uploaded files")
	rootCmd.AddCommand(uploadCmd)
}
