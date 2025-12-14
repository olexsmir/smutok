package sync

import "context"

type Strategy interface {
	Sync(ctx context.Context, initial bool) error
}
