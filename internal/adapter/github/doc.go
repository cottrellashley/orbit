// Package github implements the port.GitHubProvider interface using the
// google/go-github v84 SDK (github.com/google/go-github/v84/github).
//
// A fresh SDK client is created per API call so that token rotations
// are picked up automatically. No external state is cached between
// calls.
//
// # Token resolution
//
// The adapter resolves a personal access token using the following
// cascade (first non-empty value wins):
//
//  1. GITHUB_TOKEN environment variable
//  2. GH_TOKEN environment variable
//  3. Config file at <configDir>/github-token (single line, trimmed)
//  4. `gh auth token` command output (requires GitHub CLI)
//
// Tokens are never logged or included in error messages.
//
// # Rate limiting
//
// When the SDK returns a *github.RateLimitError the adapter maps it
// to domain.ErrRateLimited.
package github
