package sync

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"olexsmir.xyz/smutok/internal/provider"
	"olexsmir.xyz/smutok/internal/store"
)

type FreshRSS struct {
	store *store.Sqlite
	api   *provider.FreshRSS

	ot int64
}

func NewFreshRSS(store *store.Sqlite, api *provider.FreshRSS) *FreshRSS {
	return &FreshRSS{
		store: store,
		api:   api,
	}
}

func (f *FreshRSS) Sync(ctx context.Context) error {
	ot, err := f.getLastSyncTime(ctx)
	if err != nil {
		return err
	}

	f.ot = ot
	newOt := time.Now().Unix()

	// note: sync all articles once if it's initial sync

	// todo: sync pending `mark_read`, /edit-tag takes multiple &i=<id>&i=<ca>
	// todo: sync pending `mark_unread`
	// todo: sync pending `star`
	// todo: sync pending `unstar`

	if err := f.syncTags(ctx); err != nil {
		return err
	}

	if err := f.syncSubscriptions(ctx); err != nil {
		return err
	}

	if err := f.syncUnreadItems(ctx); err != nil {
		return err
	}

	if err := f.syncUnreadItemsStatuses(ctx); err != nil {
		return err
	}

	if err := f.syncStarredItems(ctx); err != nil {
		return err
	}

	if err := f.syncStarredItemStatuses(ctx); err != nil {
		return err
	}

	return f.store.SetLastSyncTime(ctx, newOt)
}

func (f *FreshRSS) getLastSyncTime(ctx context.Context) (int64, error) {
	ot, err := f.store.GetLastSyncTime(ctx)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			slog.Info("got last sync time, returning 0")
			return 0, nil
		} else {
			return 0, err
		}
	}

	slog.Info("got last sync time", "ot", ot)
	return ot, nil
}

func (f *FreshRSS) syncTags(ctx context.Context) error {
	slog.Info("syncing tags")

	tags, err := f.api.TagList(ctx)
	if err != nil {
		return err
	}

	var errs []error
	for _, tag := range tags {
		if strings.HasPrefix(tag.ID, "user/-/state/com.google/") &&
			!strings.HasSuffix(tag.ID, "/com.google/starred") {
			continue
		}

		if err := f.store.UpsertTag(ctx, tag.ID); err != nil {
			errs = append(errs, err)
		}
	}

	slog.Info("finished tag sync", "errs", errs)
	return errors.Join(errs...)
}

func (f *FreshRSS) syncSubscriptions(ctx context.Context) error {
	slog.Info("syncing subscriptions")

	subs, err := f.api.SubscriptionList(ctx)
	if err != nil {
		return err
	}

	var errs []error
	for _, sub := range subs {
		if err := f.store.UpsertSubscription(ctx, sub.ID, sub.Title, sub.URL, sub.HTMLURL); err != nil {
			errs = append(errs, err)
		}

		for _, cat := range sub.Categories {
			if !strings.Contains(cat.ID, "user/-/label") {
				continue
			}

			// NOTE: probably redundant
			// if err := g.store.UpsertTag(ctx, cat.ID); err != nil {
			// 	errs = append(errs, err)
			// }

			if err := f.store.LinkFeedWithFolder(ctx, sub.ID, cat.ID); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// delete local feeds that are no longer available on the server
	ids := make([]string, len(subs))
	for i, s := range subs {
		ids[i] = s.ID
	}

	if err := f.store.RemoveNonExistentFeeds(ctx, ids); err != nil {
		errs = append(errs, err)
	}

	slog.Info("finished subscriptions sync", "errs", errs)
	return errors.Join(errs...)
}

func (f *FreshRSS) syncUnreadItems(ctx context.Context) error {
	slog.Info("syncing unread items")

	items, err := f.api.StreamContents(ctx,
		"user/-/state/com.google/reading-list",
		"user/-/state/com.google/read",
		f.ot,
		1000)
	if err != nil {
		return err
	}

	slog.Debug("got unread items", "len", len(items))

	var errs []error
	for _, item := range items {
		if err := f.store.UpsertArticle(ctx, item.TimestampUsec, item.Origin.StreamID, item.Title, item.Content, item.Author, item.Origin.HTMLURL, int(item.Published)); err != nil {
			errs = append(errs, err)
		}
	}

	slog.Info("finished syncing unread items", "errs", errs)
	return errors.Join(errs...)
}

func (f *FreshRSS) syncUnreadItemsStatuses(ctx context.Context) error {
	slog.Info("syncing unread items ids")

	ids, err := f.api.StreamIDs(ctx,
		"user/-/state/com.google/reading-list",
		"user/-/state/com.google/read",
		1000)
	if err != nil {
		return err
	}

	slog.Debug("got unread ids", "len", len(ids), "ids", ids)
	merr := f.store.SyncReadStatus(ctx, ids)

	slog.Info("finished syncing unread items", "err", merr)
	return merr
}

func (f *FreshRSS) syncStarredItems(ctx context.Context) error {
	slog.Info("sync stared items")

	items, err := f.api.StreamContents(ctx,
		"user/-/state/com.google/starred",
		"",
		f.ot,
		1000)
	if err != nil {
		return err
	}

	slog.Debug("got starred items", "len", len(items))

	var errs []error
	for _, item := range items {
		if err := f.store.UpsertArticle(ctx, item.TimestampUsec, item.Origin.StreamID, item.Title, item.Content, item.Author, item.Origin.HTMLURL, int(item.Published)); err != nil {
			errs = append(errs, err)
		}
	}

	slog.Info("finished syncing unstarred items", "errs", errs)
	return errors.Join(errs...)
}

func (f *FreshRSS) syncStarredItemStatuses(ctx context.Context) error {
	slog.Info("syncing starred items ids")

	ids, err := f.api.StreamIDs(ctx,
		"user/-/state/com.google/starred",
		"",
		1000)
	if err != nil {
		return err
	}

	slog.Debug("got starred ids", "len", len(ids), "ids", ids)
	merr := f.store.SyncStarredStatus(ctx, ids)

	slog.Info("finished syncing unread items", "err", merr)
	return merr
}
