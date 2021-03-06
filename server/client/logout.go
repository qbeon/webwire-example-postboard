package client

import (
	"context"

	"github.com/qbeon/webwire-example-postboard/server/apisrv/api"
)

// Logout implements the ApiClient interface
func (c *apiClient) Logout(ctx context.Context) error {
	return c.Query(ctx, api.Logout, nil, nil)
}
