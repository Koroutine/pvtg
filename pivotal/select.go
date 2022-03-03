package pivotal

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

func (pt *PivotalClient) SelectStoryTBD(projectName string, label string) (Story, Project, error) {

	var story Story
	var project Project

	projects, err := pt.GetProjects()

	if err != nil {
		return Story{}, Project{}, err
	}

	templates := &promptui.SelectTemplates{
		Active:   "• {{ .Name | green }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "  {{ .Name | green }}",
	}

	if projectName != "" {

		for _, p := range projects {
			if p.Name == projectName {
				project = p
				break
			}
		}

		if project.ID == 0 {
			return Story{}, Project{}, fmt.Errorf("project '%s' not found", projectName)
		}

	} else {
		prompt := promptui.Select{
			Label:     "Select Project",
			Items:     projects,
			Templates: templates,
		}

		i, _, err := prompt.Run()

		if err != nil {
			return Story{}, Project{}, err
		}

		project = projects[i]
	}

	stories, err := project.GetStoriesTBD(label)

	if err != nil {
		return Story{}, Project{}, err
	}

	if len(stories) == 0 {
		return Story{}, Project{}, fmt.Errorf("no stories found")
	}

	templates = &promptui.SelectTemplates{
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
		return Story{}, Project{}, err
	}

	story = stories[i]

	return story, project, nil
}
