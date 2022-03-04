/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/iancoleman/strcase"
	"github.com/koroutine/pvtg/pivotal"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a story",
	Long: `Select a Pivotal Tracker story to start
	
This will create and checkout a git branch named '<story-id>_<story-name>'.

Branch name is limited to 255 characters`,
	Run: func(cmd *cobra.Command, args []string) {

		client := pivotal.NewPivotalClient(pivotalToken)
		label := cmd.Flag("label").Value.String()
		projectName := cmd.Flag("project").Value.String()

		me, err := client.Me()

		CheckIfError(err)

		project, err := client.SelectProject(projectName)
		CheckIfError(err)

		story, err := project.SelectStoryTBD(label)
		CheckIfError(err)

		name := story.Name
		maxLength := 255 - len(fmt.Sprint(story.ID))

		if len(name) > (maxLength) {
			name = story.Name[0 : maxLength-1]
		}

		branchName := fmt.Sprintf("%v_%s", story.ID, strcase.ToSnake(name))

		fmt.Println("Opening repo at", dir)
		r, err := git.PlainOpen(dir)
		CheckIfError(err)

		w, err := r.Worktree()
		CheckIfError(err)

		err = w.Checkout(&git.CheckoutOptions{
			Create: true,
			Branch: plumbing.ReferenceName(plumbing.NewBranchReferenceName(branchName)),
			Keep:   true,
		})

		if err != nil && (errors.Is(err, git.ErrBranchExists) || strings.Contains(err.Error(), "already exists")) {
			fmt.Printf("Branch '%s' already exists, good to go!\n", branchName)
		} else {
			CheckIfError(err)
		}

		fmt.Println("Switch to branch", branchName)

		story.State = pivotal.StoryStarted
		story.Owners = append(story.Owners, me.ID)
		story, err = story.Save()

		CheckIfError(err)

		fmt.Println("Started story", story.Name)

	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringP("label", "l", "", "Filter stories by label")
	startCmd.Flags().StringP("project", "p", "", "Name of project")
}
