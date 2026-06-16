package ossucs_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/ossucs-cli/ossucs"
)

const fakeReadme2 = `## Intro CS

| Courses | Duration | Effort |
| :--: | :--: | :--: |
[Introduction to Python](https://example.com/python) | 14 weeks | 10 hrs/week
[Intro to CS](https://example.com/cs) | 10 weeks | 5 hrs/week

### Core programming

[Systematic Program Design](coursepages/spd/README.md) | 13 weeks | 8 hrs/week
[Class-based Program Design](https://example.com/class) | 13 weeks | 5 hrs/week
`

func newOSSUTestClient(ts *httptest.Server) *ossucs.Client {
	cfg := ossucs.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return ossucs.NewClient(cfg)
}

func TestCourseByIndex(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeReadme2)
	}))
	defer ts.Close()

	c := newOSSUTestClient(ts)
	co, err := c.CourseByIndex(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if co.Rank != 2 {
		t.Errorf("Rank = %d, want 2", co.Rank)
	}
	if co.Title != "Intro to CS" {
		t.Errorf("Title = %q", co.Title)
	}
}

func TestCourseByIndexNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeReadme2)
	}))
	defer ts.Close()

	c := newOSSUTestClient(ts)
	_, err := c.CourseByIndex(context.Background(), 999)
	if err == nil {
		t.Error("expected error for missing index")
	}
}

func TestCoursesBySection(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeReadme2)
	}))
	defer ts.Close()

	c := newOSSUTestClient(ts)
	courses, err := c.CoursesBySection(context.Background(), "intro cs")
	if err != nil {
		t.Fatal(err)
	}
	if len(courses) != 2 {
		t.Fatalf("want 2 courses in Intro CS, got %d", len(courses))
	}
}

func TestInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeReadme2)
	}))
	defer ts.Close()

	c := newOSSUTestClient(ts)
	info, err := c.Info(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Courses != 4 {
		t.Errorf("Courses = %d, want 4", info.Courses)
	}
	if info.Sections != 2 {
		t.Errorf("Sections = %d, want 2", info.Sections)
	}
}
