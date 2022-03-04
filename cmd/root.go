/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	pivotalToken string
	cfgFile      string
	dir          string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pvtg",
	Short: "Pivotal Tracker Git helper",
	Long: `Supports a Git development flow centered around Pivotal Tracker
	
Allows you to find, start and complete tasks without leaving the comfort of your terminal`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pvtg.yaml)")
	rootCmd.PersistentFlags().StringVar(&dir, "dir", "", "Target directory")
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
		viper.SetConfigName(".pvtg")
	}

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	cobra.CheckErr(err)

	pivotalToken = viper.GetString("token")
}
