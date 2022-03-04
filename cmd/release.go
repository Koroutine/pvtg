/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/koroutine/pvtg/pivotal"
	"github.com/spf13/cobra"
)

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := pivotal.NewPivotalClient(pivotalToken)
		projectName := cmd.Flag("project").Value.String()

		project, err := client.SelectProject(projectName)
		CheckIfError(err)

		release, err := project.SelectReleases()
		CheckIfError(err)

		fmt.Println("Opening repo at", dir)
		r, err := git.PlainOpen(dir)
		CheckIfError(err)

		headRef, err := r.Head()
		CheckIfError(err)

		if !headRef.Name().IsBranch() {
			CheckIfError(errors.New("repository head is not on a branch"))
		}

		branchName := headRef.Name().Short()

		if branchName != "main" {
			CheckIfError(errors.New("repository head is not on main"))
		}

		gitCmds := [][]string{
			{"tag", "-a", "-m", fmt.Sprintf("\"%s\"", release.Description), release.Name},
			{"push", "--tags"},
		}

		err = RunGit(gitCmds)
		CheckIfError(err)

		release.State = pivotal.StoryAccepted
		release, err = release.Save()

		CheckIfError(err)
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	releaseCmd.Flags().StringP("project", "p", "", "Name of project")
}
