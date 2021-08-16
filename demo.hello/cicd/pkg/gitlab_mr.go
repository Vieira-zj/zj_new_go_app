package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// MergeRequest struct for gitlab merge request.
type MergeRequest struct {
	IID         int    `json:"iid"`
	Title       string `json:"title"`
	State       string `json:"state"`
	TargetBR    string `json:"target_branch"`
	SourceBR    string `json:"source_branch"`
	WebURL      string `json:"web_url"`
	Repo        string
	LinkedIssue string
	Err         string
}

// ToText returns merge request data as text.
func (mr *MergeRequest) ToText() string {
	items := strings.Split(mr.Repo, "/")
	repoName := items[len(items)-1]
	return fmt.Sprintf("MR:[%s],[%s:%s->%s],[%s]\n", repoName, mr.State, mr.SourceBR, mr.TargetBR, mr.Title)
}

// NewMergeRequest create a git merge request.
func NewMergeRequest(ctx context.Context, git *GitlabTool, mrURL string) (*MergeRequest, error) {
	resp, err := git.GetSingleMR(ctx, mrURL)
	if err != nil {
		return nil, err
	}

	mr := &MergeRequest{}
	err = json.Unmarshal(resp, mr)
	if err != nil {
		wrapErr := fmt.Errorf("resp [length:%d, text:%s], error: %s", len(resp), string(resp), err.Error())
		return nil, wrapErr
	}
	mr.Repo, err = getMRRepo(mrURL)
	if err != nil {
		return nil, err
	}
	return mr, nil
}

func getMRRepo(mrURL string) (string, error) {
	items := strings.Split(mrURL, "/-/")
	if len(items) != 2 {
		return "", fmt.Errorf("invalid mr url: %s", mrURL)
	}
	return strings.Replace(items[0], gitlabHost+"/", "", 1), nil
}
