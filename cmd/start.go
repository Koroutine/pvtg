/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/iancoleman/strcase"
	"github.com/koroutine/pvtg/pivotal"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a story",
	Long: `Select a Pivotal Tracker story to start
	
This will create and checkout a git branch named '<story-id>_<story-name>'.

Branch name is limited to 255 characters`,

	Run: func(cmd *cobra.Command, args []string) {

		var project pivotal.Project
		var story pivotal.Story

		client := pivotal.NewPivotalClient(pivotalToken)

		me, err := client.Me()

		CheckIfError(err)

		projects, err := client.GetProjects()

		CheckIfError(err)

		templates := &promptui.SelectTemplates{
			Active:   "• {{ .Name | green }}",
			Inactive: "  {{ .Name | cyan }}",
			Selected: "  {{ .Name | green }}",
		}

		prompt := promptui.Select{
			Label:     "Select Project",
			Items:     projects,
			Templates: templates,
		}

		i, _, err := prompt.Run()

		CheckIfError(err)

		project = projects[i]

		stories, err := project.GetStoriesTBD()

		CheckIfError(err)

		templates = &promptui.SelectTemplates{
			Active:   "• {{ .Type }} - {{ .Name | green }} ({{.Priority }})",
			Inactive: "  {{ .Type }} - {{ .Name  | cyan }} ({{.Priority }})",
			Selected: "  {{ .Type }} - {{ .Name | green }} ({{.Priority }})",
		}

		prompt = promptui.Select{
			Label:     "Select Story",
			Items:     stories,
			Templates: templates,
		}

		i, _, err = prompt.Run()

		CheckIfError(err)

		story = stories[i]

		name := story.Name
		maxLength := 255 - len(fmt.Sprint(story.ID))

		if len(name) > (maxLength) {
			name = story.Name[0 : maxLength-1]
		}

		branchName := fmt.Sprintf("%v_%s", story.ID, strcase.ToSnake(name))

		path := cmd.Flag("path").Value.String()

		fmt.Println("Opening repo at", path)

		r, err := git.PlainOpen(cmd.Flag("path").Value.String())

		CheckIfError(err)

		w, err := r.Worktree()

		CheckIfError(err)

		err = w.Checkout(&git.CheckoutOptions{
			Create: true,
			Branch: plumbing.ReferenceName(plumbing.NewBranchReferenceName(branchName)),
			Keep:   true,
		})

		if errors.Is(err, git.ErrBranchExists) || strings.Contains(err.Error(), "already exists") {
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

	startCmd.Flags().StringP("path", "p", "", "Path to git repository")
}
