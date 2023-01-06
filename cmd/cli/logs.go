package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bucheron "github.com/wesen/bucheron/pkg"
	"github.com/wesen/glazed/pkg/cli"
	"golang.org/x/sync/errgroup"
	"os"
	"syscall"
	"time"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download files from S3",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")

		// ensure directory exists
		log.Info().Msgf("Creating directory %s", output)
		err := os.MkdirAll(output, 0755)
		cobra.CheckErr(err)

		bucket := viper.GetString("bucket")
		region := viper.GetString("region")

		glob, _ := cmd.Flags().GetStringSlice("glob")
		prefix, _ := cmd.Flags().GetString("prefix")

		ctx, cancel := context.WithCancel(context.Background())

		settings := &bucheron.BucketSettings{
			Region: region,
			Bucket: bucket,
		}
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")

		filterSettings := &bucheron.ListFilter{
			Globs:  glob,
			Prefix: prefix,
		}

		var fromDate time.Time
		var toDate time.Time

		if from != "" {
			fromDate, err = bucheron.ParseDate(from)
			cobra.CheckErr(err)
			filterSettings.From = &fromDate
		}
		if to != "" {
			toDate, err = bucheron.ParseDate(to)
			cobra.CheckErr(err)
			filterSettings.To = &toDate
		}

		progressCh := make(chan bucheron.ProgressEvent)

		errGroup, ctx2 := errgroup.WithContext(ctx)
		errGroup.Go(func() error {
			for {
				select {
				case progress, ok := <-progressCh:
					if !ok {
						cancel()
						return nil
					}
					fmt.Printf("%s %f\n", progress.Step, progress.StepProgress)
				case <-ctx2.Done():
					return ctx2.Err()
				}
			}
		})

		errGroup.Go(func() error {
			return bucheron.DownloadLogs(ctx2, settings, filterSettings, output, progressCh)
		})
		errGroup.Go(func() error {
			return bucheron.CancelOnSignal(ctx2, syscall.SIGINT, cancel)
		})

		err = errGroup.Wait()
		// check if err is for canceled context
		if err != context.Canceled {
			cobra.CheckErr(err)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all logs",
	Run: func(cmd *cobra.Command, args []string) {
		bucket := viper.GetString("bucket")
		region := viper.GetString("region")

		glob, _ := cmd.Flags().GetStringSlice("glob")
		prefix, _ := cmd.Flags().GetString("prefix")

		ctx, cancel := context.WithCancel(context.Background())

		log.Debug().Msg("Getting log keys")

		settings := &bucheron.BucketSettings{
			Region: region,
			Bucket: bucket,
		}
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")

		filterSettings := &bucheron.ListFilter{
			Globs:  glob,
			Prefix: prefix,
		}

		var fromDate time.Time
		var toDate time.Time
		var err error

		if from != "" {
			fromDate, err = bucheron.ParseDate(from)
			cobra.CheckErr(err)
			filterSettings.From = &fromDate
		}
		if to != "" {
			toDate, err = bucheron.ParseDate(to)
			cobra.CheckErr(err)
			filterSettings.To = &toDate
		}

		entryCh := make(chan bucheron.LogEntry)
		gp, of, err := cli.SetupProcessor(cmd)

		gp.OutputFormatter().SetColumnOrder([]string{"fileName", "key", "date", "size", "comment", "metadata"})

		errGroup, ctx2 := errgroup.WithContext(ctx)
		errGroup.Go(func() error {
			for {
				select {
				case entry, ok := <-entryCh:
					if !ok {
						cancel()
						return nil
					}
					row := make(map[string]interface{})
					row["fileName"] = entry.FileName
					row["key"] = entry.Key
					row["comment"] = entry.Comment
					row["date"] = entry.Date
					row["size"] = entry.Size
					row["uuid"] = entry.UUID
					row["horseStaple"] = bucheron.UUIDToHorseStaple(entry.UUID)

					for k, v := range entry.Metadata {
						row[k] = v
					}
					_ = gp.ProcessInputObject(row)
				case <-ctx2.Done():
					return ctx2.Err()
				}
			}
		})

		errGroup.Go(func() error {
			return bucheron.ListLogBucketKeys(ctx2, settings, filterSettings, entryCh)
		})
		errGroup.Go(func() error {
			return bucheron.CancelOnSignal(ctx2, syscall.SIGINT, cancel)
		})

		err = errGroup.Wait()
		// check if err is for canceled context
		if err != context.Canceled {
			cobra.CheckErr(err)
		}

		s, err := of.Output()
		cobra.CheckErr(err)
		fmt.Print(s)
	},
}

func init() {
	listCmd.Flags().String("from", "", "Start date")
	listCmd.Flags().String("to", "", "End date")
	listCmd.Flags().StringSlice("glob", []string{"*.log", "*.json"}, "Glob pattern to filter keys")
	listCmd.Flags().String("prefix", "", "S3 key prefix")

	flagsDefaults := cli.NewFlagsDefaults()
	//flagsDefaults.FieldsFilter.SortColumns = true
	flagsDefaults.FieldsFilter.Filter = "key,uuid,horseStaple"
	cli.AddFlags(listCmd, flagsDefaults)

	rootCmd.AddCommand(listCmd)

	downloadCmd.Flags().String("from", "", "Start date")
	downloadCmd.Flags().String("to", "", "End date")
	downloadCmd.Flags().StringSlice("glob", []string{"*.log", "*.json"}, "Glob pattern to filter keys")
	downloadCmd.Flags().String("prefix", "", "S3 key prefix")
	downloadCmd.Flags().String("output", "out/", "Output directory")
	rootCmd.AddCommand(downloadCmd)
}
