package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"demo.hello/utils"
)

/*
gitlab rest api docs:
https://docs.gitlab.com/ee/api/merge_requests.html#get-single-mr
*/

// GitlabTool .
type GitlabTool struct {
	host  string
	token string
	http  *utils.HTTPUtils
}

// NewGitlabTool .
func NewGitlabTool() *GitlabTool {
	return &GitlabTool{
		host:  os.Getenv("GITLAB_HOST"),
		token: os.Getenv("GITLAB_TOKEN"),
		http:  utils.NewHTTPUtils(true),
	}
}

// SearchProject .
func (git *GitlabTool) SearchProject(ctx context.Context, name, namespace string) (string, error) {
	resp, err := git.get(ctx, "projects?search="+name)
	if err != nil {
		return "", err
	}

	projects := make([]interface{}, 0)
	err = json.Unmarshal(resp, &projects)
	if err != nil {
		return "", err
	}

	if len(projects) == 1 {
		project := projects[0].(map[string]interface{})
		return fmt.Sprintf("%.0f", project["id"].(float64)), nil
	}

	for _, item := range projects {
		project := item.(map[string]interface{})
		ns := project["namespace"].(map[string]interface{})
		if project["name"].(string) == name && ns["name"].(string) == namespace {
			return fmt.Sprintf("%.0f", project["id"].(float64)), nil
		}
	}
	return "", fmt.Errorf("Project [%s] not found", name)
}

// GetSingleMR .
func (git *GitlabTool) GetSingleMR(ctx context.Context, mr string) ([]byte, error) {
	items := strings.Split(mr, "/-/")
	if len(items) != 2 {
		return nil, fmt.Errorf("invalid mr url: %s", mr)
	}

	path := strings.Replace(items[0], git.host, "", 1)
	subItems := strings.Split(path, "/")
	project := subItems[len(subItems)-1]
	namespace := subItems[len(subItems)-2]
	projectID, err := git.SearchProject(ctx, project, namespace)
	if err != nil {
		return nil, err
	}

	mrID := strings.Split(items[1], "/")[1]
	reqPath := fmt.Sprintf("projects/%s/merge_requests/%s", projectID, mrID)
	return git.get(ctx, reqPath)
}

func (git *GitlabTool) get(ctx context.Context, path string) ([]byte, error) {
	url := git.host + "/api/v4/" + formatPath(path)
	headers := map[string]string{
		"PRIVATE-TOKEN": git.token,
	}
	return git.http.Get(ctx, url, headers)
}
