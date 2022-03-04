package pivotal

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

func (pt *PivotalClient) SelectProject(name string) (Project, error) {
	var project Project

	projects, err := pt.GetProjects()

	if err != nil {
		return project, err
	}

	templates := &promptui.SelectTemplates{
		Active:   "• {{ .Name | green }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "  {{ .Name | green }}",
	}

	if name != "" {

		for _, p := range projects {
			if p.Name == name {
				project = p
				break
			}
		}

		if project.ID == 0 {
			return project, fmt.Errorf("project '%s' not found", name)
		}

	} else {
		prompt := promptui.Select{
			Label:     "Select Project",
			Items:     projects,
			Templates: templates,
		}

		i, _, err := prompt.Run()

		if err != nil {
			return project, err
		}

		project = projects[i]
	}

	return project, nil
}

func (project *Project) SelectStoryTBD(label string) (Story, error) {

	var story Story

	stories, err := project.GetStoriesTBD(label)

	if err != nil {
		return story, err
	}

	if len(stories) == 0 {
		return story, fmt.Errorf("no stories found")
	}

	templates := &promptui.SelectTemplates{
		Active:   "• {{ .TypeIcon }} - {{ .Name | green }} ({{.Priority }})",
		Inactive: "  {{ .TypeIcon }} - {{ .Name | cyan }} ({{.Priority }})",
		Selected: "  {{ .TypeIcon }} - {{ .Name | green }} ({{.Priority }})",
	}

	prompt := promptui.Select{
		Label:     "Select Story",
		Items:     stories,
		Templates: templates,
	}

	i, _, err := prompt.Run()

	if err != nil {
		return story, err
	}

	story = stories[i]

	return story, nil
}

func (project *Project) SelectReleases() (Story, error) {

	var story Story

	stories, err := project.GetReleases()

	if err != nil {
		return story, err
	}

	if len(stories) == 0 {
		return story, fmt.Errorf("no releases found")
	}

	templates := &promptui.SelectTemplates{
		Active:   "• {{ .TypeIcon }} - {{ .Name | green }} ({{.Priority }})",
		Inactive: "  {{ .TypeIcon }} - {{ .Name | cyan }} ({{.Priority }})",
		Selected: "  {{ .TypeIcon }} - {{ .Name | green }} ({{.Priority }})",
	}

	prompt := promptui.Select{
		Label:     "Select Release",
		Items:     stories,
		Templates: templates,
	}

	i, _, err := prompt.Run()

	if err != nil {
		return story, err
	}

	story = stories[i]

	return story, nil
}
