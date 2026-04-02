package state

import "testing"

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
