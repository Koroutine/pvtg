/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login [token]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		viper.Set("token", args[0])
		home, err := os.UserHomeDir()
		CheckIfError(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".pvtg")

		f, err := os.Create(path.Join(home, ".pvtg.yaml"))

		if !errors.Is(err, os.ErrExist) {
			CheckIfError(err)
		}

		f.Close()

		err = viper.WriteConfig()

		CheckIfError(err)

		fmt.Println("Logged In")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

}
