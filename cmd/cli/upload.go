package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wesen/bucheron/pkg"
	"golang.org/x/sync/errgroup"
	"syscall"
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

		errGroup := errgroup.Group{}
		// Set up a signal handler to cancel the context when the user
		// presses Ctrl+C

		errGroup.Go(func() error {
			defer cancel()
			return pkg.UploadLogs(ctx, settings, data, progressCh)
		})
		errGroup.Go(func() error {
			return pkg.CancelOnSignal(ctx, syscall.SIGINT, cancel)
		})
		errGroup.Go(func() error {
			for {
				select {
				case progress := <-progressCh:
					fmt.Printf("%s: %f\n", progress.Step, progress.StepProgress)

				case <-ctx.Done():
					return nil
				}
			}
		})

		err = errGroup.Wait()
		cobra.CheckErr(err)
	},
}

func init() {
	uploadCmd.Flags().StringP("comment", "c", "", "Comment to add to the uploaded files")
	rootCmd.AddCommand(uploadCmd)
}
