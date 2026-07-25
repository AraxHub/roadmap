package roadmap

import (
	"testing"

	"roadmap/internal/domain"
)

func TestUnlockedMap(t *testing.T) {
	chain := []domain.ChainItem{
		{SubmoduleID: "a"},
		{SubmoduleID: "b"},
		{SubmoduleID: "c"},
	}

	t.Run("first always unlocked", func(t *testing.T) {
		got := unlockedMap(chain, nil)
		if !got["a"] || got["b"] || got["c"] {
			t.Fatalf("unexpected: %#v", got)
		}
	})

	t.Run("complete a unlocks b", func(t *testing.T) {
		got := unlockedMap(chain, map[string]bool{"a": true})
		if !got["a"] || !got["b"] || got["c"] {
			t.Fatalf("unexpected: %#v", got)
		}
	})

	t.Run("complete a and b unlocks c", func(t *testing.T) {
		got := unlockedMap(chain, map[string]bool{"a": true, "b": true})
		if !got["a"] || !got["b"] || !got["c"] {
			t.Fatalf("unexpected: %#v", got)
		}
	})
}
