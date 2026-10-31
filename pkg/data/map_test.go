package data

import "testing"

func TestMapCap(t *testing.T) {
	var nilMap *Map
	if got := nilMap.Cap(); got != 0 {
		t.Errorf("nil map Cap = %d, want 0", got)
	}
	if got := new(Map).Cap(); got != 0 {
		t.Errorf("zero map Cap = %d, want 0", got)
	}
	if got := NewMap(8).Cap(); got != 8 {
		t.Errorf("NewMap(8) Cap = %d, want 8", got)
	}
	if got := NewMap(-1).Cap(); got != 0 {
		t.Errorf("NewMap(-1) Cap = %d, want 0", got)
	}

	// Cap is what a builder compares with Len to see that Append will
	// reallocate; after growth it covers Len.
	m := NewMap(2)
	for i := 0; i < 3; i++ {
		if i == 2 && m.Len() != m.Cap() {
			t.Fatalf("Len %d != Cap %d before the append that reallocates", m.Len(), m.Cap())
		}
		m.Append("k", Value{})
	}
	if m.Cap() <= 2 || m.Cap() < m.Len() {
		t.Errorf("after growth Cap = %d, Len = %d", m.Cap(), m.Len())
	}
}
