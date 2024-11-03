package constants

// ContextKey is used to represent typed constant strings to be used in keys of context.Context.
type ContextKey string

const (
	// UserContext is a gin context key in which the user information stored on gin.Context.
	UserContext = "userProfile"
)

const (
	// Session string constants.
	Session = "Session"
	// Authorization header key.
	Authorization = "Authorization"
	// Cookie header key.
	Cookie = "Cookie"
	// Bearer constant.
	Bearer = "Bearer"
	// NullString used to check for empty string.
	NullString = ""
	// FakeOperationID Dummy OperationID constant.
	FakeOperationID = "fake-operation-id"
	// TraceID W3 trace id.
	TraceID = "trace-id"
	// Dependency value is log type for logger.
	Dependency = "dependency"
	// Source is key stored in context to preserve the source system call triggered from.
	Source = "source"
	// Target is destination service is being called.
	Target = "target"
	// GRPC is gRPC protocol.
	GRPC = "gRPC"
	// HTTP is a string constant for using in dependency operations.
	HTTP = "HTTP"
	// TCP is a string constant for using in dependency operations.
	TCP = "TCP"
	// Error is a string constant to be used in error level logs as a key.
	Error = "error"
)
