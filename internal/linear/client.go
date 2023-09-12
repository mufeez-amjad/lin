package linear

import (
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
	httpClient := http.Client{
		Transport: &authedTransport{
			key:     "lin_api_CWyCx6GctBCvCyxRYIDLA45g06XZuNgh5GBsqco3",
			wrapped: http.DefaultTransport,
		},
	}
	graphqlClient = graphql.NewClient("http://localhost:8090/graphql", &httpClient)
}

func GetClient() GqlClient {
	initClientOnce.Do(func() {
		initClient()
	})
	return graphqlClient
}
