package linear

import (
	"context"
	"encoding/json"
	"lin_cli/internal/store"
	"log"
)

type Issue struct {
	Id          string
	Identifier  string
	Title       string
	Description string
	BranchName  string
	Url         string
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
			assignedIssues(after: $cursor, orderBy: updatedAt, filter: {
				state: {
				  type: {
					in: ["started", "backlog"]
				  }
				}
			}) {
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
				}
			}
		}
	}
	`

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
			issue := &Issue{
				Id:          issue.Id,
				Identifier:  issue.Identifier,
				Title:       issue.Title,
				Description: issue.Description,
				BranchName:  issue.BranchName,
				Url:         issue.Url,
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
