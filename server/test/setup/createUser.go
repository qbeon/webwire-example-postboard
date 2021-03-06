package setup

import (
	"context"
	"time"

	"github.com/qbeon/webwire-example-postboard/server/apisrv/api"
	"github.com/qbeon/webwire-example-postboard/server/client"
	"github.com/stretchr/testify/require"
)

// CreateUser creates a user profile expecting the operation
// to be successful, retrieves the profile and verifies whether all fields
// of the profile are correct
func (h *Helper) CreateUser(
	admin client.ApiClient,
	params api.CreateUserParams,
) (*api.User, client.ApiClient) {
	// Create and expect no errors
	newUserIdent, err := admin.CreateUser(
		context.Background(),
		api.CreateUserParams{
			FirstName: params.FirstName,
			LastName:  params.LastName,
			Username:  params.Username,
			Password:  params.Password,
			Type:      params.Type,
		},
	)
	require.NoError(h.t, err)
	require.False(h.t, newUserIdent.IsNull())

	// Log in to the newly created account
	var clt client.ApiClient
	if params.Type == api.UtAdmin {
		clt = h.ts.NewAdminClient(params.Username, params.Password)
	} else if params.Type == api.UtUser {
		clt = h.ts.NewUserClient(params.Username, params.Password)
	} else {
		h.ts.t.Errorf("unsupported user type: %s", params.Type)
		return nil, nil
	}

	// Verify session identifier
	session := clt.Session()
	require.NotNil(h.t, session.Info.Value("id"))
	require.IsType(h.t, api.Identifier{}, session.Info.Value("id"))
	sessionIdent := session.Info.Value("id").(api.Identifier)
	require.NotNil(h.t, session)
	require.Equal(h.t,
		newUserIdent.String(),
		sessionIdent.String(),
	)

	// Retrieve profile
	profile, err := clt.GetUser(context.Background(), api.GetUserParams{
		Ident: *clt.Identifier(),
	})
	require.NoError(h.t, err)
	require.NotNil(h.t, profile)

	// Verify profile information
	require.Equal(h.t, params.FirstName, profile.FirstName)
	require.Equal(h.t, params.LastName, profile.LastName)
	require.Equal(h.t, params.Username, profile.Username)
	require.Equal(h.t, params.Type.String(), profile.Type.String())
	require.Equal(h.t, float64(0), profile.Reputation)
	require.WithinDuration(h.t,
		time.Now().UTC(),
		profile.Registration,
		h.ts.MaxCreationTimeDeviation(),
	)

	return profile, clt
}
