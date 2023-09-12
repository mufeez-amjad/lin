package linear

import (
	"context"
	"encoding/json"
	"io"
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

func (i Issue) Serialize(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(i)
}

func (i Issue) Deserialize(r io.Reader) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(&i)
}

func GetIssues(client GqlClient) ([]Issue, error) {
	_ = `# @genqlient
	query getAssignedIssues {
	  viewer {
		assignedIssues {
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

	issues := []Issue{}
	objs := []store.Serializable{}

	for {
		issuesResp, err := getAssignedIssues(context.Background(), graphqlClient)
		if err != nil {
			return nil, err
		}

		assignedIssuesConnection := issuesResp.Viewer.AssignedIssues
		for _, issue := range assignedIssuesConnection.Nodes {
			issue := Issue{
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
		// cursor = pageInfo.EndCursor
	}

	err := store.WriteObjectToFile("./cache", objs)
	if err != nil {
		log.Fatalf("Failed to write to cache")
	}

	return issues, nil
}
