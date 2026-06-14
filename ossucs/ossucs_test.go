package ossucs_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/ossucs-cli/ossucs"
)

const fakeReadme = `## Intro CS

| Courses | Duration | Effort |
| :--: | :--: | :--: |
[Introduction to Python](https://example.com/python) | 14 weeks | 10 hrs/week

### Core programming

[Systematic Program Design](coursepages/spd/README.md) | 13 weeks | 8 hrs/week
[Class-based Program Design](https://example.com/class) | 13 weeks | 5 hrs/week
`

func newTestClient(ts *httptest.Server) *ossucs.Client {
	cfg := ossucs.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return ossucs.NewClient(cfg)
}

func TestCourses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeReadme)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	courses, err := c.Courses(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(courses) != 3 {
		t.Fatalf("want 3, got %d", len(courses))
	}
	if courses[0].Section != "Intro CS" {
		t.Errorf("Section[0] = %q", courses[0].Section)
	}
	if courses[0].Title != "Introduction to Python" {
		t.Errorf("Title[0] = %q", courses[0].Title)
	}
	if courses[1].Section != "Core programming" {
		t.Errorf("Section[1] = %q", courses[1].Section)
	}
	if courses[1].URL != "https://github.com/ossu/computer-science/blob/master/coursepages/spd/README.md" {
		t.Errorf("URL[1] = %q", courses[1].URL)
	}
	if courses[0].Rank != 1 {
		t.Errorf("Rank = %d, want 1", courses[0].Rank)
	}
}
