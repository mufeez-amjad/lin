package linear

import (
	"context"
	"encoding/json"
	"fmt"
	"lin_cli/internal/store"
	"log"
	"time"

	"github.com/Khan/genqlient/graphql"
)

type Organization struct {
	Teams []*Team
}

type Team struct {
	Id     string
	Key    string
	Name   string
	Color  string
	States []*State
}

type State struct {
	Id       string
	Name     string
	Color    string
	Type     string
	Position int
}

func (i *Organization) Serialize() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Organization) Deserialize(data []byte) error {
	err := json.Unmarshal(data, i)
	return err
}

func getTeams(client GqlClient) {

}

func GetOrganization(client GqlClient) (*Organization, error) {
	_ = `# @genqlient
query getOrganization(
	# @genqlient(omitempty: true)
	$cursor: String
) {
  viewer {
    organization {
      teams(after: $cursor) {
        pageInfo {
          endCursor
          hasNextPage
        }
        nodes {
          id
          key
          name
          color
        }
      }
    }
  }
}
`
	organization := &Organization{
		Teams: []*Team{},
	}

	cursor := ""
	for {
		orgsResp, err := getOrganization(context.Background(), graphqlClient, cursor)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve organization, %s", err)
		}

		teamsConnection := orgsResp.Viewer.Organization.Teams
		for _, team := range teamsConnection.Nodes {
			organization.Teams = append(organization.Teams, &Team{
				Id:    team.Id,
				Name:  team.Name,
				Key:   team.Key,
				Color: team.Color,
			})
		}

		pageInfo := teamsConnection.PageInfo
		if !pageInfo.HasNextPage {
			break
		}
		cursor = pageInfo.EndCursor
	}

	for _, team := range organization.Teams {
		if err := team.GetTeamStatusStates(client); err != nil {
			return nil, fmt.Errorf("Failed to retrieve team states: %s", err)
		}
	}

	objs := []store.Serializable{
		organization,
	}
	err := store.WriteObjectToFile("./org.cache", objs)
	if err != nil {
		log.Fatalf("Failed to write to cache")
	}

	return organization, nil
}

func (t *Team) GetTeamStatusStates(client GqlClient) error {
	_ = `# @genqlient
query teamStates(
	$teamId: String!
	# @genqlient(omitempty: true)
	$cursor: String
) {
  team(id: $teamId) {
	states(after: $cursor) {
      pageInfo {
        endCursor
        hasNextPage
      }
      nodes {
        id
        name
        color
        type
        position
      }
    }
  }
}
`
	cursor := ""
	for {
		statesResp, err := teamStates(context.Background(), graphqlClient, t.Id, cursor)
		if err != nil {
			return err
		}

		statesConnection := statesResp.Team.States

		for _, state := range statesConnection.Nodes {
			t.States = append(t.States, &State{
				Id:       state.Id,
				Name:     state.Name,
				Color:    state.Color,
				Type:     state.Type,
				Position: int(state.Position),
			})
		}

		pageInfo := statesConnection.PageInfo
		if !pageInfo.HasNextPage {
			break
		}
		cursor = pageInfo.EndCursor
	}

	return nil
}

// Retrieves issues from cache
func LoadOrg(client graphql.Client) (org *Organization, needRefresh bool, err error) {
	var lastCached time.Time
	orgs, lastCached, err := store.ReadObjectFromFile[*Organization]("./org.cache", func() *Organization {
		return &Organization{}
	})

	if len(orgs) != 0 {
		org = orgs[0]
	}

	isFresh := lastCached.Add(time.Hour * 12).Before(time.Now())
	needRefresh = isFresh || org == nil

	return org, needRefresh, err
}
