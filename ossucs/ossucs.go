// Package ossucs is the library behind the ossucs command line:
// the HTTP client, request shaping, and the typed data models for the OSSU
// Computer Science curriculum.
//
// The Client here is the spine every command shares. It sets a real
// User-Agent, paces requests so a busy session stays polite, and retries the
// transient failures (429 and 5xx) that any public site throws under load.
package ossucs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// DefaultUserAgent identifies the client to the remote host.
const DefaultUserAgent = "ossucs/dev (+https://github.com/tamnd/ossucs-cli)"

// Host is the site the URI driver in domain.go claims.
const Host = "github.com/ossu/computer-science"

var (
	headingRe = regexp.MustCompile(`^#{1,3}\s+(.+)`)
	courseRe  = regexp.MustCompile(`^\[([^\]]+)\]\(([^)]+)\)\s*\|`)
	ghBase    = "https://github.com/ossu/computer-science/blob/master/"
)

// Config holds all tunables for the Client.
type Config struct {
	BaseURL   string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
	UserAgent string
}

// DefaultConfig returns a Config pointed at raw.githubusercontent.com with
// conservative pacing and retry settings.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://raw.githubusercontent.com",
		Rate:      200 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   5,
		UserAgent: DefaultUserAgent,
	}
}

// Client talks to raw.githubusercontent.com over HTTP.
type Client struct {
	cfg  Config
	HTTP *http.Client
	last time.Time
}

// NewClient returns a Client configured by cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		HTTP: &http.Client{Timeout: cfg.Timeout},
	}
}

// Courses fetches the OSSU Computer Science README and returns all courses
// with their section, title, URL, and rank.
func (c *Client) Courses(ctx context.Context) ([]*Course, error) {
	body, err := c.get(ctx, "/ossu/computer-science/master/README.md")
	if err != nil {
		return nil, err
	}
	var courses []*Course
	section := ""
	rank := 1
	for _, line := range strings.Split(string(body), "\n") {
		if m := headingRe.FindStringSubmatch(line); m != nil {
			section = strings.TrimSpace(m[1])
			continue
		}
		m := courseRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := m[1]
		url := m[2]
		if !strings.HasPrefix(url, "http") {
			url = ghBase + url
		}
		courses = append(courses, &Course{
			Rank:    rank,
			Section: section,
			Title:   name,
			URL:     url,
		})
		rank++
	}
	if len(courses) == 0 {
		return nil, fmt.Errorf("no courses found")
	}
	return courses, nil
}

// get fetches a path under BaseURL and returns the response body. It paces and
// retries according to the client's settings.
func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	url := c.cfg.BaseURL + path
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has passed since the previous request.
func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}
