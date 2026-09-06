package testutil

// Port mocks are generated with gomock's reflect mode. Regenerate with
// `go generate ./internal/testutil/...` after changing any interface in
// internal/core/ports.
//
//go:generate go tool mockgen -destination mocks.go -package testutil github.com/ViitoJooj/foreman/internal/core/ports TaskRepository,CompanyRepository,AgentRepository,ChannelRepository,CommandRepository,MessageRepository,MessageBus,GitHubClient,BrowserClient
