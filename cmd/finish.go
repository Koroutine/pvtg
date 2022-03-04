/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

		fmt.Println("Opening repo at", dir)
		r, err := git.PlainOpen(dir)
		CheckIfError(err)

		w, err := r.Worktree()
		CheckIfError(err)

		headRef, err := r.Head()
		CheckIfError(err)

		if !headRef.Name().IsBranch() {
			CheckIfError(errors.New("repository head is not on a branch"))
		}

		branchName := headRef.Name().Short()
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
		status, err := w.Status()
		CheckIfError(err)

		if !status.IsClean() {
			CheckIfError(fmt.Errorf("branch '%s' has uncommitted changes", branchName))
		}

		// Checkout dev
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName(plumbing.NewBranchReferenceName("dev")),
			Force:  true,
		})

		if errors.Is(err, git.ErrBranchNotFound) {
			CheckIfError(errors.New("dev branch not found"))
			// TODO prompt to make dev branch
		} else {
			CheckIfError(err)
		}

		gitCmds := [][]string{
			{"git", "merge", branchName},
			{"git", "push"},
			{"git", "branch", "-d", branchName},
			{"git", "push", "origin", "--delete", branchName},
		}

		for _, c := range gitCmds {
			// Pull changes from target branch
			cmd := exec.Command(c[0], c[1:]...)
			cmd.Dir = dir
			err = cmd.Run()
			CheckIfError(err)
		}

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
