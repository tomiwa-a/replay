package state

import (
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	s := NewStore()

	s.Set("foo", "bar")
	val, ok := s.Get("foo")
	if !ok || val != "bar" {
		t.Errorf("expected bar, got %v", val)
	}

	s.Set("count", 42)
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 items, got %d", len(all))
	}
}

func TestStoreScopedGet(t *testing.T) {
	s := NewStore()

	s.Set("global_key", "global_val")

	s.Enter(ScopeWorkflow)
	s.Set("wf_key", "wf_val")

	v, ok := s.Get("global_key")
	if !ok || v != "global_val" {
		t.Errorf("expected global_key=global_val, got %v", v)
	}

	v, ok = s.Get("wf_key")
	if !ok || v != "wf_val" {
		t.Errorf("expected wf_key=wf_val, got %v", v)
	}

	s.Enter(ScopeStep)
	s.Set("step_key", "step_val")

	v, ok = s.Get("global_key")
	if !ok || v != "global_val" {
		t.Errorf("global not visible in step scope")
	}
	v, ok = s.Get("wf_key")
	if !ok || v != "wf_val" {
		t.Errorf("workflow not visible in step scope")
	}
	v, ok = s.Get("step_key")
	if !ok || v != "step_val" {
		t.Errorf("step key not visible in step scope")
	}

	s.Exit()
	if _, ok := s.Get("step_key"); ok {
		t.Errorf("step key should not be visible after Exit()")
	}

	s.ExitUntil(ScopeGlobal)
	if _, ok := s.Get("wf_key"); ok {
		t.Errorf("wf key should not be visible after ExitUntil(global)")
	}
}

func TestStoreScopeWriteVisibility(t *testing.T) {
	s := NewStore()
	s.Set("key", "global")

	s.Enter(ScopeWorkflow)
	s.Set("key", "workflow")
	v, _ := s.Get("key")
	if v != "workflow" {
		t.Errorf("expected workflow, got %v", v)
	}

	s.Exit()
	v, _ = s.Get("key")
	if v != "global" {
		t.Errorf("expected global after exit, got %v", v)
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	s.Set("foo", "bar")

	s.Delete("foo")
	if _, ok := s.Get("foo"); ok {
		t.Error("expected foo to be deleted")
	}

	s.Enter(ScopeWorkflow)
	s.Set("bar", "baz")
	s.Delete("bar")
	if _, ok := s.Get("bar"); ok {
		t.Error("expected bar to be deleted from workflow scope")
	}

	s.Exit()
}

func TestStoreSnapshotRestore(t *testing.T) {
	s := NewStore()
	s.Set("a", 1)
	s.Set("b", 2)

	snap := s.Snapshot()
	s.Set("c", 3)

	v, ok := s.Get("c")
	if !ok || v != 3 {
		t.Error("c should be visible before restore")
	}

	s.Restore(snap)
	if _, ok := s.Get("c"); ok {
		t.Error("c should not be visible after restore")
	}
	v, _ = s.Get("a")
	if v != 1 {
		t.Error("a should still be 1")
	}
}

func TestStoreTTL(t *testing.T) {
	s := NewStore()
	s.SetTTL("temp", "value", 50*time.Millisecond)

	v, ok := s.Get("temp")
	if !ok || v != "value" {
		t.Error("temp should be visible immediately")
	}

	time.Sleep(80 * time.Millisecond)
	if _, ok := s.Get("temp"); ok {
		t.Error("temp should have expired")
	}

	snap := s.All()
	if _, ok := snap["temp"]; ok {
		t.Error("expired key should not appear in All()")
	}
}

func TestStoreAllFlattensScopes(t *testing.T) {
	s := NewStore()
	s.Set("a", 1)

	s.Enter(ScopeWorkflow)
	s.Set("b", 2)

	s.Enter(ScopeStep)
	s.Set("c", 3)

	all := s.All()
	if len(all) != 3 {
		t.Errorf("expected 3 items, got %d: %v", len(all), all)
	}
	if all["a"] != 1 || all["b"] != 2 || all["c"] != 3 {
		t.Errorf("unexpected values in All(): %v", all)
	}
}

func TestStoreOverwriteAcrossScopes(t *testing.T) {
	s := NewStore()
	s.Set("key", "global")

	s.Enter(ScopeWorkflow)
	s.Set("key", "workflow")
	v, _ := s.Get("key")
	if v != "workflow" {
		t.Errorf("expected workflow, got %v", v)
	}

	s.Exit()
	v, _ = s.Get("key")
	if v != "global" {
		t.Errorf("expected global after exit, got %v", v)
	}
}

func TestStoreScopeDepth(t *testing.T) {
	s := NewStore()
	if s.ScopeDepth() != 1 {
		t.Errorf("expected depth 1, got %d", s.ScopeDepth())
	}

	s.Enter(ScopeWorkflow)
	if s.ScopeDepth() != 2 {
		t.Errorf("expected depth 2, got %d", s.ScopeDepth())
	}

	s.Enter(ScopeStep)
	if s.ScopeDepth() != 3 {
		t.Errorf("expected depth 3, got %d", s.ScopeDepth())
	}

	s.Exit()
	if s.ScopeDepth() != 2 {
		t.Errorf("expected depth 2 after exit, got %d", s.ScopeDepth())
	}
}

func TestStoreMustGet(t *testing.T) {
	s := NewStore()
	s.Set("exists", "yes")

	if v := s.MustGet("exists"); v != "yes" {
		t.Errorf("expected yes, got %v", v)
	}

	if v := s.MustGet("nope"); v != nil {
		t.Errorf("expected nil, got %v", v)
	}
}
