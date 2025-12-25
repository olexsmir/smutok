package freshrss

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"olexsmir.xyz/smutok/internal/store"
)

type Syncer struct {
	store *store.Sqlite
	api   *Client

	ot int64
}

func NewSyncer(api *Client, store *store.Sqlite) *Syncer {
	return &Syncer{
		store: store,
		api:   api,
	}
}

func (f *Syncer) Sync(ctx context.Context) error {
	ot, err := f.getLastSyncTime(ctx)
	if err != nil {
		return err
	}

	f.ot = ot
	newOt := time.Now().Unix()

	// TODO: sync all articles once if it's initial sync

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

func (f *Syncer) getLastSyncTime(ctx context.Context) (int64, error) {
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

func (f *Syncer) syncTags(ctx context.Context) error {
	slog.Info("syncing tags")

	tags, err := f.api.TagList(ctx)
	if err != nil {
		return err
	}

	var errs []error
	for _, tag := range tags {
		if strings.HasPrefix(tag.ID, "user/-/state/com.google/") &&
			!strings.HasSuffix(tag.ID, StateStarred) {
			continue
		}

		if err := f.store.UpsertTag(ctx, tag.ID); err != nil {
			errs = append(errs, err)
		}
	}

	slog.Info("finished tag sync", "errs", errs)
	return errors.Join(errs...)
}

func (f *Syncer) syncSubscriptions(ctx context.Context) error {
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

func (f *Syncer) syncUnreadItems(ctx context.Context) error {
	slog.Info("syncing unread items")

	items, err := f.api.StreamContents(ctx, StreamContents{
		StreamID:      StateReadingList,
		ExcludeTarget: StateRead,
		LastModified:  f.ot,
		N:             1000,
	})
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

func (f *Syncer) syncUnreadItemsStatuses(ctx context.Context) error {
	slog.Info("syncing unread items ids")

	ids, err := f.api.StreamIDs(ctx, StreamID{
		IncludeTarget: StateReadingList,
		ExcludeTarget: StateRead,
		N:             1000,
	})
	if err != nil {
		return err
	}

	slog.Debug("got unread ids", "len", len(ids), "ids", ids)
	merr := f.store.SyncReadStatus(ctx, ids)

	slog.Info("finished syncing unread items", "err", merr)
	return merr
}

func (f *Syncer) syncStarredItems(ctx context.Context) error {
	slog.Info("sync stared items")

	items, err := f.api.StreamContents(ctx, StreamContents{
		StreamID:     StateStarred,
		LastModified: f.ot,
		N:            1000,
	})
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

func (f *Syncer) syncStarredItemStatuses(ctx context.Context) error {
	slog.Info("syncing starred items ids")

	ids, err := f.api.StreamIDs(ctx, StreamID{
		IncludeTarget: StateStarred,
		N:             1000,
	})
	if err != nil {
		return err
	}

	slog.Debug("got starred ids", "len", len(ids), "ids", ids)
	merr := f.store.SyncStarredStatus(ctx, ids)

	slog.Info("finished syncing unread items", "err", merr)
	return merr
}
