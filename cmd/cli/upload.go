package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wesen/bucheron/pkg"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
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

		credentials, err := pkg.GetUploadCredentials()
		cobra.CheckErr(err)

		settings := &pkg.UploadSettings{
			Region:      region,
			Bucket:      bucket,
			Credentials: credentials,
		}

		data := &pkg.UploadData{
			Files:    args,
			Comment:  comment,
			Metadata: nil,
		}

		ctx, cancel := context.WithCancel(context.Background())

		// Set up a signal handler to cancel the context when the user
		// presses Ctrl+C
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT)
		go func() {
			<-sigCh
			fmt.Println("Received Ctrl+C, canceling context")
			cancel()
		}()

		progressCh := make(chan pkg.UploadProgress)

		errGroup := errgroup.Group{}
		errGroup.Go(func() error {
			defer cancel()
			return pkg.UploadLogs(ctx, settings, data, progressCh)
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
