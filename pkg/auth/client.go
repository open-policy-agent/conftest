package auth

import (
	"context"
	"errors"

	"github.com/containerd/containerd/remotes"
)

// Common errors
var (
	ErrNotLoggedIn = errors.New("not logged in")
)

// Client provides authentication operations for remotes.
type Client interface {
	// Resolver returns a new authenticated resolver.
	Resolver(ctx context.Context) (remotes.Resolver, error)
}
