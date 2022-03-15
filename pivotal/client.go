package pivotal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type StoryState string
type StoryType string

const (
	StoryAll         StoryState = "all"
	StoryAccepted    StoryState = "accepted"
	StoryDelivered   StoryState = "delivered"
	StoryFinished    StoryState = "finished"
	StoryStarted     StoryState = "started"
	StoryRejected    StoryState = "rejected"
	StoryPlanned     StoryState = "planned"
	StoryUnstarted   StoryState = "unstarted"
	StoryUnscheduled StoryState = "unscheduled"
)

const (
	StoryAny     StoryType = "any"
	StoryRelease StoryType = "release"
	StoryChore   StoryType = "chore"
	StoryBug     StoryType = "bug"
	StoryFeature StoryType = "feature"
)

type Me struct {
	ID       int    `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

type Project struct {
	ID   int            `json:"id,omitempty"`
	Name string         `json:"name,omitempty"`
	pt   *PivotalClient `json:"-"`
}

type Story struct {
	ID          int            `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Type        StoryType      `json:"story_type,omitempty"`
	TypeIcon    string         `json:"-"`
	State       StoryState     `json:"current_state,omitempty"`
	Owners      []int          `json:"owner_ids,omitempty"`
	Priority    string         `json:"story_priority,omitempty"`
	Estimate    float32        `json:"estimate,omitempty"`
	ProjectID   int            `json:"-"`
	pt          *PivotalClient `json:"-"`
}

type PivotalClient struct {
	client *http.Client
	token  string
}

func NewPivotalClient(token string) *PivotalClient {

	client := http.DefaultClient

	return &PivotalClient{
		client,
		token,
	}
}

func (pt *PivotalClient) Me() (Me, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.pivotaltracker.com/services/v5/me", http.NoBody)

	if err != nil {
		return Me{}, err
	}

	req.Header.Set("X-TrackerToken", pt.token)

	res, err := pt.client.Do(req)

	if err != nil {
		return Me{}, err
	}

	var me Me

	err = json.NewDecoder(res.Body).Decode(&me)

	if err != nil {
		return Me{}, err
	}

	return me, nil
}

func (pt *PivotalClient) GetProjects() ([]Project, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.pivotaltracker.com/services/v5/projects", http.NoBody)

	if err != nil {
		return nil, err
	}

	req.Header.Set("X-TrackerToken", pt.token)

	res, err := pt.client.Do(req)

	if err != nil {
		return nil, err
	}

	var projects []Project

	err = json.NewDecoder(res.Body).Decode(&projects)

	if err != nil {
		return nil, err
	}

	sort.Slice(projects, func(i, j int) bool {

		return strings.Compare(projects[i].Name, projects[j].Name) == 1
	})

	for i := range projects {
		projects[i].pt = pt
	}

	return projects, nil
}

func (pt *PivotalClient) GetProjectByName(name string) (Project, error) {

	var project Project
	projects, err := pt.GetProjects()

	if err != nil {
		return project, err
	}

	for _, p := range projects {
		if p.Name == name {
			project = p
			break
		}
	}

	if project.ID == 0 {
		return project, fmt.Errorf("project '%s' not found", name)
	}

	return project, nil
}

func (project *Project) GetStories(state StoryState, label string) ([]Story, error) {
	url, err := url.Parse(fmt.Sprintf("https://www.pivotaltracker.com/services/v5/projects/%v/stories", project.ID))

	if err != nil {
		return nil, err
	}

	values := url.Query()

	if state != StoryAll {
		values.Set("with_state", string(state))
	}
	if label != "" {
		values.Set("with_label", label)
	}

	url.RawQuery = values.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), http.NoBody)

	if err != nil {
		return nil, err
	}

	req.Header.Set("X-TrackerToken", project.pt.token)

	res, err := project.pt.client.Do(req)

	if err != nil {
		return nil, err
	}

	var stories []Story

	err = json.NewDecoder(res.Body).Decode(&stories)

	if err != nil {
		return nil, err
	}

	sort.Slice(stories, func(i, j int) bool {

		priority := strings.Compare(stories[i].Priority, stories[j].Priority)

		if priority == 1 {
			return true
		} else if priority == -1 {
			return false
		} else {
			storyType := strings.Compare(string(stories[i].Type), string(stories[j].Type))

			return storyType == 1
		}
	})

	for i := range stories {
		stories[i].pt = project.pt
		stories[i].ProjectID = project.ID

		switch stories[i].Type {
		case StoryFeature:
			stories[i].TypeIcon = "⭐"
		case StoryBug:
			stories[i].TypeIcon = "🐛"
		case StoryChore:
			stories[i].TypeIcon = "🍩"
		case StoryRelease:
			stories[i].TypeIcon = "🏁"
		}
	}

	return stories, nil

}

func (project *Project) GetStory(id int) (Story, error) {

	var story Story

	url := fmt.Sprintf("https://www.pivotaltracker.com/services/v5/projects/%v/stories/%v", project.ID, id)
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)

	if err != nil {
		return story, err
	}

	req.Header.Set("X-TrackerToken", project.pt.token)

	res, err := project.pt.client.Do(req)

	if err != nil {
		return story, err
	}

	err = json.NewDecoder(res.Body).Decode(&story)

	if err != nil {
		return story, err
	}

	story.pt = project.pt
	story.ProjectID = project.ID

	switch story.Type {
	case StoryFeature:
		story.TypeIcon = "⭐"
	case StoryBug:
		story.TypeIcon = "🐛"
	case StoryChore:
		story.TypeIcon = "🍩"
	case StoryRelease:
		story.TypeIcon = "🏁"
	}

	return story, nil
}

func (project *Project) GetStoriesTBD(label string) ([]Story, error) {
	states := []StoryState{
		StoryUnstarted,
		StoryRejected,
	}

	stories := make([]Story, 0)

	for _, state := range states {
		data, err := project.GetStories(state, label)

		if err != nil {
			return nil, err
		}

		for _, d := range data {

			hasEstimate := d.Estimate != 0

			if d.Type == StoryChore {
				hasEstimate = true
			}

			if d.Type != StoryRelease && hasEstimate {
				stories = append(stories, d)
			}
		}
	}

	return stories, nil
}

func (project *Project) GetReleases() ([]Story, error) {

	stories := make([]Story, 0)

	data, err := project.GetStories(StoryUnstarted, "")

	if err != nil {
		return nil, err
	}

	for _, d := range data {
		if d.Type == StoryRelease {
			stories = append(stories, d)
		}
	}

	return stories, nil
}

func (story *Story) Save() (Story, error) {

	url := fmt.Sprintf("https://www.pivotaltracker.com/services/v5/projects/%v/stories/%v", story.ProjectID, story.ID)

	body, err := json.Marshal(Story{
		State:  story.State,
		Owners: story.Owners,
	})

	if err != nil {
		return *story, err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))

	if err != nil {
		return *story, err
	}

	req.Header.Set("X-TrackerToken", story.pt.token)
	req.Header.Set("Content-Type", "application/json")

	res, err := story.pt.client.Do(req)

	if err != nil {
		return *story, err
	}

	if res.StatusCode != http.StatusOK {
		body := make(map[string]interface{})

		err = json.NewDecoder(res.Body).Decode(&body)

		if err != nil {
			return *story, errors.New(res.Status)
		}

		errorMsg, ok := body["general_problem"]
		if !ok {
			return *story, errors.New(res.Status)
		}

		err = errors.New(errorMsg.(string))
		if err != nil {
			return *story, err
		}
	}

	updateStory := make(map[string]interface{})

	err = json.NewDecoder(res.Body).Decode(&updateStory)

	if err != nil {
		return *story, err
	}

	return *story, nil
}
