package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "cobra-cli",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Error executing root command")
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "author name for copyright attribution")
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	if err := viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author")); err != nil {
		log.Fatal().Err(err).Msg("Error binding author flag")
	}
	if err := viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper")); err != nil {
		log.Fatal().Err(err).Msg("Error binding viper flag")
	}
	viper.SetDefault("author", "Ashish Bhatia")

	urlSuggestCmd.Flags().StringVarP(&HugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	urlSuggestCmd.Flags().BoolVarP(&UpdateInline, "in-place", "", false, "Update titles in in markdown files")
	urlSuggestCmd.PersistentFlags().BoolVarP(&ColorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	rootCmd.AddCommand(urlSuggestCmd)

	siteSummaryCmd.Flags().StringVarP(&HugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	siteSummaryCmd.PersistentFlags().BoolVarP(&ColorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	rootCmd.AddCommand(siteSummaryCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
