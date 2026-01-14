package tgbot

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/gotd/td/tg"
)

// albumCollector collects grouped messages (albums) and fires a callback
// when the album is complete.
type albumCollector struct {
	mu       sync.Mutex
	albums   map[int64][]*tg.Message // groupedID -> messages
	timers   map[int64]*time.Timer
	timeout  time.Duration
	callback func(ctx context.Context, messages []*tg.Message)
}

func newAlbumCollector(timeout time.Duration, callback func(ctx context.Context, messages []*tg.Message)) *albumCollector {
	return &albumCollector{
		albums:   make(map[int64][]*tg.Message),
		timers:   make(map[int64]*time.Timer),
		timeout:  timeout,
		callback: callback,
	}
}

// add adds a message to the album collector.
// Returns true if the message was added to an album (should not be processed individually).
func (c *albumCollector) add(ctx context.Context, msg *tg.Message) bool {
	if msg.GroupedID == 0 {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	groupID := msg.GroupedID
	c.albums[groupID] = append(c.albums[groupID], msg)

	// Cancel existing timer
	if timer, ok := c.timers[groupID]; ok {
		timer.Stop()
	}

	// Set new timer
	c.timers[groupID] = time.AfterFunc(c.timeout, func() {
		c.flush(ctx, groupID)
	})

	return true
}

// flush processes and removes an album.
func (c *albumCollector) flush(ctx context.Context, groupID int64) {
	c.mu.Lock()
	messages := c.albums[groupID]
	delete(c.albums, groupID)
	delete(c.timers, groupID)
	c.mu.Unlock()

	if len(messages) == 0 {
		return
	}

	// Sort by message ID to ensure correct order
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})

	c.callback(ctx, messages)
}

// stop cancels all pending album timers.
func (c *albumCollector) stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, timer := range c.timers {
		timer.Stop()
	}
	c.albums = make(map[int64][]*tg.Message)
	c.timers = make(map[int64]*time.Timer)
}
