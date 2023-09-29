package linear

import (
	"context"
	"encoding/json"
	"lin_cli/internal/store"
	"log"
	"time"

	"github.com/Khan/genqlient/graphql"
)

type Issue struct {
	Id          string
	Identifier  string
	Title       string
	Description string
	BranchName  string
	Url         string
	Attachments []*Attachment
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
      filter: { state: { type: { in: ["started", "backlog"] } } }
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

	err := store.WriteObjectToFile("./cache", objs)
	if err != nil {
		log.Fatalf("Failed to write to cache")
	}

	return issues, nil
}

// Retrieves issues from cache
func LoadIssues(client graphql.Client) (issues []*Issue, needRefresh bool, err error) {
	var lastCached time.Time
	issues, lastCached, err = store.ReadObjectFromFile[*Issue]("./cache", func() *Issue {
		return &Issue{}
	})

	isFresh := lastCached.Add(time.Hour * 12).Before(time.Now())
	needRefresh = isFresh || len(issues) == 0

	return issues, needRefresh, err
}
