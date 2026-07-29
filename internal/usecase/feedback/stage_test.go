package feedback

import (
	"testing"

	"roadmap/internal/domain"
)

func TestResolveStage(t *testing.T) {
	chain := []domain.ChainItem{
		{SubmoduleID: "a", SubmoduleTitle: "A1", ModuleTitle: "M1", SprintTitle: "S1"},
		{SubmoduleID: "b", SubmoduleTitle: "B1", ModuleTitle: "M1", SprintTitle: "S1"},
		{SubmoduleID: "c", SubmoduleTitle: "C1", ModuleTitle: "M2", SprintTitle: "S2"},
	}

	t.Run("empty chain", func(t *testing.T) {
		status, stage := resolveStage(nil, nil)
		if status != "empty" || stage != nil {
			t.Fatalf("got %s %#v", status, stage)
		}
	})

	t.Run("first unlocked incomplete", func(t *testing.T) {
		status, stage := resolveStage(chain, nil)
		if status != "in_progress" || stage == nil || stage.SubmoduleTitle != "A1" {
			t.Fatalf("got %s %#v", status, stage)
		}
	})

	t.Run("second after first completed", func(t *testing.T) {
		status, stage := resolveStage(chain, map[string]bool{"a": true})
		if status != "in_progress" || stage == nil || stage.SubmoduleTitle != "B1" {
			t.Fatalf("got %s %#v", status, stage)
		}
	})

	t.Run("all completed", func(t *testing.T) {
		status, stage := resolveStage(chain, map[string]bool{"a": true, "b": true, "c": true})
		if status != "completed" || stage != nil {
			t.Fatalf("got %s %#v", status, stage)
		}
	})
}
