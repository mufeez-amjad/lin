package linear

import (
	"lin_cli/internal/config"
	"net/http"
	"sync"

	"github.com/Khan/genqlient/graphql"
)

type GqlClient = graphql.Client

var (
	graphqlClient  GqlClient
	initClientOnce sync.Once
)

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

func initClient() {
	config := config.GetConfig()

	httpClient := http.Client{
		Transport: &authedTransport{
			key:     config.APIKey,
			wrapped: http.DefaultTransport,
		},
	}
	graphqlClient = graphql.NewClient(config.GraphQLEndpoint, &httpClient)
}

func GetClient() GqlClient {
	initClientOnce.Do(func() {
		initClient()
	})
	return graphqlClient
}
