package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/xanzy/go-gitlab"
)

// GetAPIToken returns the token from the CI environment
func getAPIToken() string {
	if os.Getenv("COMPOSER_MR_TOKEN") != "" {
		return os.Getenv("COMPOSER_MR_TOKEN")
	}
	// fallback to original variable
	return os.Getenv("GITLAB_API_PRIVATE_TOKEN")
}

// Client will set up the Gitlab API connection
func client() (*gitlab.Client, error) {
	if gitClient != nil {
		return gitClient, nil
	}

	token := getAPIToken()
	projectID := os.Getenv("CI_PROJECT_ID")
	projectPath := os.Getenv("CI_PROJECT_PATH")
	repositoryURL := os.Getenv("CI_REPOSITORY_URL")
	apiURL := os.Getenv("CI_API_V4_URL")
	if token == "" || projectID == "" || projectPath == "" || repositoryURL == "" || apiURL == "" {
		return nil, fmt.Errorf("gitlab environment variables not set")
	}

	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(apiURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return git, nil
}

// IsMREnabled is a simple test to determine whether merge requests are enabled for a project.
func isMREnabled() bool {
	client, err := client()
	if err != nil {
		fmt.Println("Error authenticating with API: ", err)
		return false
	}

	opts := gitlab.ListProjectMergeRequestsOptions{
		Search: gitlab.String("Just a random string that won't return too many results (*^$%#"),
	}

	_, _, err = client.MergeRequests.ListProjectMergeRequests(os.Getenv("CI_PROJECT_ID"), &opts)
	if err != nil {
		fmt.Println("Error listing MRs: ", err)
		return false
	}

	return true
}

// MRExists checks to see if an existing merge request exists
// based on checksum of the content
func MRExists(checksum string) bool {
	client, err := client()
	if err != nil {
		fmt.Println("Error authenticating with API: ", err)
		return false
	}

	opts := gitlab.ListProjectMergeRequestsOptions{
		State:        gitlab.String("opened"),
		TargetBranch: gitlab.String(Config.GitBranch),
		Labels:       envCSVSlice("COMPOSER_MR_LABELS", []string{}),
	}

	me, _, err := client.Users.CurrentUser()
	if err == nil {
		opts.AuthorID = &me.ID
	}

	mrs, _, err := client.MergeRequests.ListProjectMergeRequests(os.Getenv("CI_PROJECT_ID"), &opts)
	if err != nil {
		fmt.Println("Error listing MRs: ", err)
		return false
	}

	for _, mr := range mrs {
		if strings.Contains(mr.Description, checksum) {
			return true
		}
	}

	return false
}

// RemoveOldMRs will remove old merge requests (if enabled)
// by deleting the branches
func RemoveOldMRs() error {
	if !envTrue("COMPOSER_MR_REPLACE_OPEN", true) {
		return nil
	}

	client, err := client()
	if err != nil {
		return fmt.Errorf("error authenticating with API: %s", err)
	}

	opts := gitlab.ListProjectMergeRequestsOptions{
		State:        gitlab.String("opened"),
		TargetBranch: gitlab.String(Config.GitBranch),
		Labels:       envCSVSlice("COMPOSER_MR_LABELS", []string{}),
	}

	me, _, err := client.Users.CurrentUser()
	if err == nil {
		opts.AuthorID = &me.ID
	}

	mrs, _, err := client.MergeRequests.ListProjectMergeRequests(os.Getenv("CI_PROJECT_ID"), &opts)
	if err != nil {
		return fmt.Errorf("error listing MRs: %s", err)
	}

	for _, mr := range mrs {
		if strings.HasPrefix(mr.Title, "Composer update: ") {
			if err := deleteOriginBranch(mr.SourceBranch); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateMergeRequest will create a merge request for the branch
// setting the title, description and other options
func CreateMergeRequest(title, description string) error {
	client, err := client()
	if err != nil {
		return err
	}

	opts := gitlab.CreateMergeRequestOptions{
		Title:              gitlab.String(title),
		Description:        gitlab.String(description),
		SourceBranch:       gitlab.String(Config.MRBranch),
		TargetBranch:       gitlab.String(Config.GitBranch),
		RemoveSourceBranch: gitlab.Bool(true),
		AssigneeIDs:        getAssigneeIDS(),
		ReviewerIDs:        getReviewerIDS(),
		Labels:             envCSVSlice("COMPOSER_MR_LABELS", []string{}),
	}
	mr, _, err := client.MergeRequests.CreateMergeRequest(os.Getenv("CI_PROJECT_ID"), &opts)
	if err != nil {
		return err
	}

	fmt.Printf("\n==========\nMerge request #%d created: %s\n", mr.ID, mr.WebURL)

	if len(mr.Labels) > 0 {
		fmt.Println("Labels:", strings.Join(mr.Labels, ", "))
	}

	if len(mr.Assignees) > 0 {
		fmt.Println("Assigned to:")
		for _, a := range mr.Assignees {
			fmt.Println("-", a.Username)
		}
	}

	if len(mr.Reviewers) > 0 {
		fmt.Println("Reviewers assigned:")
		for _, a := range mr.Reviewers {
			fmt.Println("-", a.Username)
		}
	}

	fmt.Println("==========")

	return nil
}

// GetAssigneeIDS returns a slice of IDs assigned to the new merge request
func getAssigneeIDS() []int {
	assigneeIDs := []int{}

	client, err := client()
	if err != nil {
		return assigneeIDs
	}

	assignees := envCSVSlice("COMPOSER_MR_ASSIGNEES", []string{})

	if len(assignees) > 0 {
		lookup := make(map[string]string, len(assignees))
		for _, m := range assignees {
			lookup[strings.ToLower(strings.TrimSpace(m))] = m
		}
		members, _, err := client.ProjectMembers.ListAllProjectMembers(os.Getenv("CI_PROJECT_ID"), &gitlab.ListProjectMembersOptions{})
		if err == nil {
			for _, m := range members {
				if _, ok := lookup[strings.ToLower(m.Username)]; ok {
					assigneeIDs = append(assigneeIDs, m.ID)
				}
			}
		}
	}

	return assigneeIDs
}

// GetReviewerIDS returns a slice of IDs assigned to the new merge request
func getReviewerIDS() []int {
	reviewerIDs := []int{}

	client, err := client()
	if err != nil {
		return reviewerIDs
	}

	reviewers := envCSVSlice("COMPOSER_MR_REVIEWERS", []string{})

	if len(reviewers) > 0 {
		lookup := make(map[string]string, len(reviewers))
		for _, m := range reviewers {
			lookup[strings.ToLower(strings.TrimSpace(m))] = m
		}
		members, _, err := client.ProjectMembers.ListAllProjectMembers(os.Getenv("CI_PROJECT_ID"), &gitlab.ListProjectMembersOptions{})
		if err == nil {
			for _, m := range members {
				if _, ok := lookup[strings.ToLower(m.Email)]; ok {
					reviewerIDs = append(reviewerIDs, m.ID)
				}
				if _, ok := lookup[strings.ToLower(m.Username)]; ok {
					reviewerIDs = append(reviewerIDs, m.ID)
				}
			}
		}
	}

	return reviewerIDs
}
