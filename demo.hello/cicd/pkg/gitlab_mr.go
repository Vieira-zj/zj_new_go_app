package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// MergeRequest gitlab merge request info.
type MergeRequest struct {
	IID      int    `json:"iid"`
	Title    string `json:"title"`
	State    string `json:"state"`
	TargetBR string `json:"target_branch"`
	SourceBR string `json:"source_branch"`
	WebURL   string `json:"web_url"`
	Repo     string
	Err      string
}

// PrintText prints merge request info as text.
func (mr *MergeRequest) PrintText(prefix string) {
	fmt.Printf("%s%s", prefix, mr.ToText())
}

// ToText returns merge request info as text.
func (mr *MergeRequest) ToText() string {
	return fmt.Sprintf("MR:[%s:%s->%s],[%s]\n", mr.State, mr.SourceBR, mr.TargetBR, mr.Title)
}

// NewMergeRequest create a merge request instance.
func NewMergeRequest(ctx context.Context, git *GitlabTool, mrURL string) (*MergeRequest, error) {
	resp, err := git.GetSingleMR(ctx, mrURL)
	if err != nil {
		return nil, err
	}

	mr := &MergeRequest{}
	err = json.Unmarshal(resp, mr)
	if err != nil {
		return nil, err
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
	return strings.Replace(items[0], os.Getenv("GITLAB_HOST")+"/", "", 1), nil
}
