package sync

import (
	"context"

	"olexsmir.xyz/smutok/internal/provider"
	"olexsmir.xyz/smutok/internal/store"
)

type FreshRSS struct {
	store store.Store
	api   *provider.FreshRSS
}

func NewFreshRSS(store store.Store, api *provider.FreshRSS) *FreshRSS {
	return &FreshRSS{
		store: store,
		api:   api,
	}
}

func (g *FreshRSS) Sync(ctx context.Context, initial bool) error {
	writeToken, err := g.api.GetWriteToken(ctx)
	if err != nil {
		return err
	}

	_ = writeToken

	return nil
}
