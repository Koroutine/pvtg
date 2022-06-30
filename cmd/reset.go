/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bitfield/script"
	"github.com/koroutine/pvtg/pivotal"
	"github.com/spf13/cobra"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Return story to Unstarted state",
	Long:  `This will return a PivotalTracker story to 'Unstarted', remove you from the Owners, and checkout 'dev' branch`,
	Run: func(cmd *cobra.Command, args []string) {
		client := pivotal.NewPivotalClient(pivotalToken)
		projectName := cmd.Flag("project").Value.String()

		project, err := client.SelectProject(projectName)
		CheckIfError(err)

		me, err := client.Me()
		CheckIfError(err)

		fmt.Println("Using repo at", dir)

		branchName, err := script.Exec("git rev-parse --abbrev-ref HEAD").String()
		CheckIfError(err)

		branchName = strings.TrimSpace(branchName)

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

		fmt.Println("Checking out dev")

		gitCmds := [][]string{
			{"checkout", "dev"},
		}

		err = RunGit(gitCmds)
		CheckIfError(err)

		// Update pivotal status

		owners := make([]int, 0)
		for _, o := range story.Owners {
			if o != me.ID {
				owners = append(owners, o)
			}
		}

		story.State = pivotal.StoryUnstarted
		story.Owners = owners
		_, err = story.Save()
		CheckIfError(err)

		fmt.Printf("Reset story %v\n", id)
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
	resetCmd.Flags().StringP("project", "p", "", "Name of project")
}
