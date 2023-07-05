package main

import (
	"embed"
	"fmt"
	"github.com/go-go-golems/bucheron/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "bucheron",
	Short: "bucheron CLI tool",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("verbose") {
			fmt.Println("Verbose output enabled")
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
	},
}

func initViper() error {
	viper.SetEnvPrefix("bucheron")

	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.bucheron")

	// Read the configuration file into Viper
	err := viper.ReadInConfig()
	// if the file does not exist, continue normally
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found; ignore error
	} else if err != nil {
		// Config file was found but another error was produced
		return err
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Bind the variables to the command-line flags
	err = viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		return err
	}

	return nil
}

var getCredentialsCommand = &cobra.Command{
	Use:   "get-credentials",
	Short: "Get temporary credentials for uploading file",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		credentials, err := pkg.GetUploadCredentials(ctx, viper.GetString("api"))
		cobra.CheckErr(err)

		gp, _, err := cli.CreateGlazedProcessorFromCobra(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not create glaze  procersors: %v\n", err)
			os.Exit(1)
		}

		row := types.NewRow(
			types.MRP("access_key_id", credentials.AccessKeyID),
			types.MRP("secret_access_key", credentials.SecretAccessKey),
			types.MRP("session_token", credentials.SessionToken),
			types.MRP("expiration", credentials.Expiration.Format(time.RFC3339)),
		)
		_ = gp.AddRow(ctx, row)

		err = gp.Close(ctx)
		cobra.CheckErr(err)
	},
}

func main() {
	_ = rootCmd.Execute()
}

//go:embed doc/*
var docFS embed.FS

func init() {
	helpSystem := help.NewHelpSystem()

	err := helpSystem.LoadSectionsFromFS(docFS, ".")
	if err != nil {
		panic(err)
	}

	helpSystem.SetupCobraRootCommand(rootCmd)
	rootCmd.PersistentFlags().StringP(
		"api", "a",
		"https://npyksyvjqj.execute-api.us-east-1.amazonaws.com/v1/",
		"URL of the bucheron API")
	rootCmd.PersistentFlags().StringP(
		"bucket", "b",
		"wesen-ppa-control-logs",
		"S3 bucket to upload to")
	rootCmd.PersistentFlags().StringP(
		"region", "r",
		"us-east-1",
		"Region of the S3 bucket")

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")

	helpCmd := help.NewCobraHelpCommand(helpSystem)
	rootCmd.SetHelpCommand(helpCmd)

	err = cli.AddGlazedProcessorFlagsToCobraCommand(getCredentialsCommand)
	cobra.CheckErr(err)

	rootCmd.AddCommand(getCredentialsCommand)

	err = initViper()
	cobra.CheckErr(err)
}
