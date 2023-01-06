package main

import (
	"embed"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wesen/bucheron/pkg"
	"github.com/wesen/glazed/pkg/cli"
	"github.com/wesen/glazed/pkg/help"
	"os"
	"strings"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "bucheron",
	Short: "bucheron CLI tool",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
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
		credentials, err := pkg.GetUploadCredentials()
		cobra.CheckErr(err)

		gp, of, err := cli.SetupProcessor(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not create glaze  procersors: %v\n", err)
			os.Exit(1)
		}

		row := map[string]interface{}{
			"access_key_id":     credentials.AccessKeyID,
			"secret_access_key": credentials.SecretAccessKey,
			"session_token":     credentials.SessionToken,
			"expiration":        credentials.Expiration.Format(time.RFC3339),
		}
		_ = gp.ProcessInputObject(row)

		s, err := of.Output()
		cobra.CheckErr(err)
		fmt.Print(s)
	},
}

func main() {
	_ = rootCmd.Execute()
}

//go:embed doc/*
var docFS embed.FS

func init() {
	helpSystem := help.NewHelpSystem()

	err := helpSystem.LoadSectionsFromEmbedFS(docFS, ".")
	if err != nil {
		panic(err)
	}

	helpFunc, usageFunc := help.GetCobraHelpUsageFuncs(helpSystem)
	helpTemplate, usageTemplate := help.GetCobraHelpUsageTemplates(helpSystem)

	_ = usageFunc
	_ = usageTemplate

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

	rootCmd.SetHelpFunc(helpFunc)
	rootCmd.SetUsageFunc(usageFunc)
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)

	helpCmd := help.NewCobraHelpCommand(helpSystem)
	rootCmd.SetHelpCommand(helpCmd)

	cli.AddOutputFlags(getCredentialsCommand)
	cli.AddTemplateFlags(getCredentialsCommand)
	cli.AddFieldsFilterFlags(getCredentialsCommand, "")
	cli.AddSelectFlags(getCredentialsCommand)
	cli.AddRenameFlags(getCredentialsCommand)

	rootCmd.AddCommand(getCredentialsCommand)

	initViper()
}
