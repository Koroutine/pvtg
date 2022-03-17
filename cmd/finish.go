/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/bitfield/script"
	"github.com/koroutine/pvtg/pivotal"
	"github.com/spf13/cobra"
)

// finishCmd represents the finish command
var finishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Merge current branch and Finish story",
	Long: `Finish current branch and story
	
This will find the current story from the branch name '<story-id>_<story-name>',
merge into 'dev' branch and finish the story`,
	Run: func(cmd *cobra.Command, args []string) {
		client := pivotal.NewPivotalClient(pivotalToken)
		projectName := cmd.Flag("project").Value.String()

		project, err := client.SelectProject(projectName)
		CheckIfError(err)

		fmt.Println("Using repo at", dir)

		branchName, err := script.Exec("git rev-parse --abbrev-ref HEAD").String()
		CheckIfError(err)

		if branchName == "HEAD" {
			CheckIfError(errors.New("repository head is not on a branch"))
		}

		re := regexp.MustCompile(`^(\d{9})_`)
		result := re.FindStringSubmatch(branchName)

		if len(result) != 2 {
			CheckIfError(fmt.Errorf("current branch is not a story: %s", branchName))
		}

		id, err := strconv.Atoi(result[1])
		CheckIfError(err)

		story, err := project.GetStory(id)
		CheckIfError(err)

		// Get repo status
		changes, err := script.Exec("git status --porcelain").String()
		CheckIfError(err)

		if changes != "" {
			CheckIfError(fmt.Errorf("branch '%s' has uncommitted changes", branchName))
		}

		gitCmds := [][]string{
			{"checkout", "dev"},
			{"merge", branchName},
			{"push"},
			{"branch", "-d", branchName},
			{"push", "origin", "--delete", branchName},
		}

		err = RunGit(gitCmds)
		CheckIfError(err)

		// Update pivotal status
		story.State = pivotal.StoryFinished
		_, err = story.Save()
		CheckIfError(err)

		fmt.Printf("Merged %s into dev and Finished story\n", branchName)
	},
}

func init() {
	rootCmd.AddCommand(finishCmd)
	finishCmd.Flags().StringP("project", "p", "", "Name of project")
}
