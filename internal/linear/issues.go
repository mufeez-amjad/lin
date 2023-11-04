package linear

import (
	"context"
	"encoding/json"
	"lin_cli/internal/git"
	"lin_cli/internal/store"
	"log"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/cli/go-gh/v2/pkg/api"
)

type Issue struct {
	Id          string
	Identifier  string
	Title       string
	Description string
	BranchName  string
	Url         string
	Attachments []*Attachment
	State       IssueState
}

type IssueState struct {
	Id    string
	Name  string
	Color string
}

type Attachment struct {
	Title     string
	Subtitle  string
	Url       string
	UpdatedAt time.Time
}

func (i *Issue) Serialize() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Issue) Deserialize(data []byte) error {
	err := json.Unmarshal(data, i)
	return err
}

type GitStatus int

const (
	None GitStatus = iota
	HasBranch
	HasPR
)

func getPRStatus() {
	client, err := api.DefaultRESTClient()
	if err != nil {
		log.Fatal(err)
	}
	response := []struct {
		Name string
	}{}
	err = client.Get("repos/cli/cli/tags", &response)
	if err != nil {
		log.Fatal(err)
	}
}

func (i *Issue) GetGitStatus() GitStatus {
	branches, err := git.FindBranches(i.Identifier)
	if err != nil {
		return None
	}

	hasBranch := len(branches) > 0
	hasPR := len(i.Attachments) > 0

	if hasBranch {
		return HasBranch
	}
	if hasPR {
		return HasPR
	}

	return None
}

func GetIssues(client GqlClient) ([]*Issue, error) {
	_ = `# @genqlient
query getAssignedIssues(
  # @genqlient(omitempty: true)
  $cursor: String
) {
  viewer {
    assignedIssues(
      after: $cursor
      orderBy: updatedAt
      filter: { state: { type: { in: ["backlog", "unstarted", "started"] } } }
    ) {
      pageInfo {
        hasNextPage
        endCursor
      }
      nodes {
        id
        identifier
        title
        description
        branchName
        url
		state {
			id
			name
			color
		}
        attachments(filter: { sourceType: { in: ["github", "gitlab"] } }) {
          nodes {
            title
            subtitle
            url
            updatedAt
          }
        }
      }
    }
  }
}`

	issues := []*Issue{}
	objs := []store.Serializable{}

	cursor := ""
	for {
		issuesResp, err := getAssignedIssues(context.Background(), graphqlClient, cursor)
		if err != nil {
			return nil, err
		}

		assignedIssuesConnection := issuesResp.Viewer.AssignedIssues
		for _, issue := range assignedIssuesConnection.Nodes {

			attachments := []*Attachment{}
			for _, attachment := range issue.Attachments.Nodes {
				attachments = append(attachments, &Attachment{
					Title:     attachment.Title,
					Subtitle:  attachment.Subtitle,
					Url:       attachment.Url,
					UpdatedAt: attachment.UpdatedAt,
				})
			}

			issue := &Issue{
				Id:          issue.Id,
				Identifier:  issue.Identifier,
				Title:       issue.Title,
				Description: issue.Description,
				BranchName:  issue.BranchName,
				Url:         issue.Url,
				Attachments: attachments,
				State: IssueState{
					Id:    issue.State.Id,
					Name:  issue.State.Name,
					Color: issue.State.Color,
				},
			}
			issues = append(issues, issue)
			objs = append(objs, issue)
		}

		pageInfo := assignedIssuesConnection.PageInfo
		if !pageInfo.HasNextPage {
			break
		}
		cursor = pageInfo.EndCursor
	}

	err := store.WriteObjectToFile("./issues.cache", objs)
	if err != nil {
		log.Fatalf("Failed to write to cache")
	}

	return issues, nil
}

// Retrieves issues from cache
func LoadIssues(client graphql.Client) (issues []*Issue, needRefresh bool, err error) {
	var lastCached time.Time
	issues, lastCached, err = store.ReadObjectFromFile[*Issue]("./issues.cache", func() *Issue {
		return &Issue{}
	})

	isFresh := lastCached.Add(time.Hour * 12).Before(time.Now())
	needRefresh = isFresh || len(issues) == 0

	return issues, needRefresh, err
}
