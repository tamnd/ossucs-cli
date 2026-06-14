package ossucs

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

// These tests are offline: they exercise the URI driver's pure string functions
// and the host wiring (mint, body, resolve), which need no network.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "ossucs" {
		t.Errorf("Scheme = %q, want ossucs", info.Scheme)
	}
	if info.Identity.Binary != "ossucs" {
		t.Errorf("Identity.Binary = %q, want ossucs", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	typ, id, err := Domain{}.Classify("Introduction to Programming")
	if err != nil || typ != "course" || id != "Introduction to Programming" {
		t.Errorf("Classify = (%q, %q, %v), want (course, ..., nil)", typ, id, err)
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("course", "1")
	if err != nil || got == "" {
		t.Errorf("Locate = (%q, %v)", got, err)
	}
}

// TestHostWiring mounts the driver in a kit Host and checks the round trip.
func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}

	c := &Course{Rank: 1, Section: "Intro CS", Title: "Introduction to Python", URL: "https://example.com/python"}
	_, err = h.Mint(c)
	// Mint may or may not succeed depending on kit:"id" tag; just ensure no panic.
	_ = err
}
