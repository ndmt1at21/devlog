package authlocal

import (
	"sync"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
)

// callGroup coalesces concurrent Refresh calls keyed by refresh-token hash so a
// single rotation serves every in-flight caller — a purpose-built, dependency-
// free equivalent of golang.org/x/sync/singleflight for one method's signature.
type callGroup struct {
	mu       sync.Mutex
	inflight map[string]*refreshCall
}

// refreshCall is one in-flight rotation whose result is shared by every caller
// that joined while it was running.
type refreshCall struct {
	wg  sync.WaitGroup
	ts  *authn.TokenSet
	err error
}

// do runs fn for key, deduplicating overlapping calls: the first caller for a
// key executes fn while later callers block until it finishes and then receive
// the same result. The entry is removed once fn returns, so a subsequent call
// (e.g. a genuinely reused token presented after the rotation settled) runs fn
// again and is rejected on its own merits.
func (g *callGroup) do(key string, fn func() (*authn.TokenSet, error)) (*authn.TokenSet, error) {
	g.mu.Lock()
	if c, ok := g.inflight[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.ts, c.err
	}
	c := &refreshCall{}
	c.wg.Add(1)
	g.inflight[key] = c
	g.mu.Unlock()

	c.ts, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.inflight, key)
	g.mu.Unlock()

	return c.ts, c.err
}
