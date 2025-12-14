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

func (g *FreshRSS) Sync(ctx context.Context) error {
	// tags, err := g.api.TagList(ctx)
	// subscriptions, err := g.api.SubscriptionList(ctx)
	// unreadItems, err := g.api.StreamContents(
	// 	ctx,
	// 	"user/-/state/com.google/reading-list",
	// 	"user/-/state/com.google/read",
	// 	0,
	// 	1000)
	// ids, err := g.api.GetItemsIDs(ctx,
	// 	"user/-/state/com.google/read",
	// 	"user/-/state/com.google/reading-list",
	// 	1000,
	// )

	return nil
}
