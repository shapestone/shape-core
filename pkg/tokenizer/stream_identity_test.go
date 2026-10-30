package tokenizer

import (
	"strings"
	"testing"
)

func TestStreamMatchIdentity(t *testing.T) {
	constructors := map[string]func() Stream{
		"string":   func() Stream { return NewStream("abc") },
		"buffered": func() Stream { return NewStreamFromReader(strings.NewReader("abc")) },
	}

	for name, newStream := range constructors {
		t.Run(name+"/clone matches", func(t *testing.T) {
			s := newStream()
			clone := s.Clone()
			clone.NextChar()
			s.Match(clone) // must not panic
		})

		t.Run(name+"/clone of clone matches", func(t *testing.T) {
			s := newStream()
			s.Match(s.Clone().Clone())
		})

		t.Run(name+"/different streams panic", func(t *testing.T) {
			a, b := newStream(), newStream()
			defer func() {
				if got := recover(); got != "trying to match two different streams" {
					t.Fatalf("recover() = %v, want mismatch panic", got)
				}
			}()
			a.Match(b)
		})
	}
}
