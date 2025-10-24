package main

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
	"testing"
)

func TestMessageIsStory(t *testing.T) {
	tc := []struct {
		n   string
		exp bool
	}{
		{n: "story", exp: true},
		{n: "text", exp: false},
	}
	for _, tt := range tc {
		p := filepath.Join("testdata", tt.n+".json")
		t.Run(p, func(t *testing.T) {
			f, err := os.Open(p)
			if err != nil {
				t.Errorf("expect no error opening %s, got %s", p, err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					t.Errorf("expected no error closing %s, got %s", p, err)
				}
			}()
			var u update
			if err = json.UnmarshalRead(f, &u); err != nil {
				t.Errorf("expect no error decoding %s, got %s", p, err)
			}
			got := u.isStory()
			if got != tt.exp {
				t.Errorf("expected isStory to return %t for %s, got %t", tt.exp, p, got)
			}
		})
	}
}
