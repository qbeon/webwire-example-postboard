package api

// Method represents an API method name
type Method string

const (
	// Login represents an API method
	Login Method = "Login"

	// Logout represents an API method
	Logout Method = "Logout"

	// CreateUser represents an API method
	CreateUser Method = "CreateUser"

	// GetUser represents an API method
	GetUser Method = "GetUser"

	// GetPost represents an API method
	GetPost Method = "GetPost"

	// GetPosts represents an API method
	GetPosts Method = "GetPosts"

	// GetPostReaction represents an API method
	GetPostReaction = "GetPostReaction"

	// GetReactionsOfPost represents an API method
	GetReactionsOfPost = "GetReactionsOfPost"

	// CreatePost represents an API method
	CreatePost Method = "CreatePost"

	// EditPost represents an API method
	EditPost Method = "EditPost"

	// RemovePost represents an API method
	RemovePost Method = "RemovePost"

	// RemovePostReaction represents an API method
	RemovePostReaction Method = "RemovePostReaction"

	// CreatePostReaction represents an API method
	CreatePostReaction Method = "CreatePostReaction"
)
