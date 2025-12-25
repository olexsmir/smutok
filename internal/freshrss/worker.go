package freshrss

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"olexsmir.xyz/smutok/internal/store"
)

type Worker struct {
	api   *Client
	store *store.Sqlite

	writeToken string
}

func NewWorker(api *Client, store *store.Sqlite, writeToken string) *Worker {
	return &Worker{
		api:        api,
		store:      store,
		writeToken: writeToken,
	}
}

func (w *Worker) Run(ctx context.Context) {
	// TODO: get tick time from config ???
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !w.isNetworkAvailable(ctx) {
				slog.Info("worker: no internet connection")
				continue
			}

			wg.Go(func() {
				if err := w.pendingReads(ctx); err != nil {
					slog.Error("worker: read", "err", err)
				}
			})
			wg.Go(func() {
				if err := w.pendingUnreads(ctx); err != nil {
					slog.Error("worker: unread", "err", err)
				}
			})
			wg.Go(func() {
				if err := w.pendingStar(ctx); err != nil {
					slog.Error("worker: star", "err", err)
				}
			})
			wg.Go(func() {
				if err := w.pendingUnstar(ctx); err != nil {
					slog.Error("worker: unread", "err", err)
				}
			})
			wg.Wait()
		}
	}
}

// TODO: implement me
func (Worker) isNetworkAvailable(_ context.Context) bool {
	return true
}

func (w *Worker) pendingReads(ctx context.Context) error {
	slog.Debug("worker: pending read")
	return w.handle(ctx, store.Read, StateRead, "")
}

func (w *Worker) pendingUnreads(ctx context.Context) error {
	slog.Debug("worker: pending unread")
	return w.handle(ctx, store.Unread, StateKeptUnread, StateRead)
}

func (w *Worker) pendingStar(ctx context.Context) error {
	slog.Debug("worker: pending star")
	return w.handle(ctx, store.Star, StateStarred, "")
}

func (w *Worker) pendingUnstar(ctx context.Context) error {
	slog.Debug("worker: pending unstar")
	return w.handle(ctx, store.Unstar, "", StateStarred)
}

func (w *Worker) handle(ctx context.Context, action store.Action, addState, rmState string) error {
	articleIDs, err := w.store.GetPendingActions(ctx, action)
	if err != nil {
		return err
	}

	if len(articleIDs) == 0 {
		return nil
	}

	if err := w.api.EditTag(ctx, w.writeToken, EditTag{
		ItemID:      articleIDs,
		TagToAdd:    addState,
		TagToRemove: rmState,
	}); err != nil {
		return err
	}

	return w.store.DeletePendingActions(ctx, action, articleIDs)
}
