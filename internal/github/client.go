// Package github fetches profile data from the GitHub GraphQL API.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const endpoint = "https://api.github.com/graphql"

// Client issues authenticated GraphQL requests.
type Client struct {
	token string
	http  *http.Client
}

// NewClient returns a client authenticated with the given PAT. An empty token
// falls back to unauthenticated requests (60/h rate limit, no private data).
func NewClient(token string) *Client {
	return &Client{
		token: token,
		http:  &http.Client{Timeout: 30 * time.Second},
	}
}

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type gqlError struct {
	Message string   `json:"message"`
	Type    string   `json:"type,omitempty"`
	Path    []string `json:"path,omitempty"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlError      `json:"errors,omitempty"`
}

// maxRateLimitSleep caps how long we're willing to wait for a rate-limit
// reset before giving up — a 1-hour reset window is better handled by the
// caller (reschedule the Action) than by sleeping through it.
const maxRateLimitSleep = 5 * time.Minute

// query runs a GraphQL query and unmarshals the `data` field into out.
// Respects ctx deadlines so pagination loops can abort early when the
// caller's overall budget expires. On a primary-rate-limit 403, honors
// Retry-After / X-RateLimit-Reset once before retrying.
func (c *Client) query(ctx context.Context, q string, vars map[string]any, out any) error {
	body, err := json.Marshal(gqlRequest{Query: q, Variables: vars})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("new request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "ghstats")
		if c.token != "" {
			req.Header.Set("Authorization", "bearer "+c.token)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("http: %w", err)
		}

		raw, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read body: %w", err)
		}

		if rateLimited(resp) && attempt == 0 {
			wait := rateLimitWait(resp)
			if wait > maxRateLimitSleep {
				return fmt.Errorf("http %d: rate limit resets in %s (>%s max wait)", resp.StatusCode, wait, maxRateLimitSleep)
			}
			fmt.Fprintf(os.Stderr, "warn: rate-limited, sleeping %s before retry\n", wait.Round(time.Second))
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		if resp.StatusCode >= 400 {
			return fmt.Errorf("http %d: %s", resp.StatusCode, truncate(raw, 500))
		}

		var r gqlResponse
		if err := json.Unmarshal(raw, &r); err != nil {
			return fmt.Errorf("decode body: %w", err)
		}
		if len(r.Errors) > 0 {
			msgs := make([]string, 0, len(r.Errors))
			for _, e := range r.Errors {
				msgs = append(msgs, e.Message)
			}
			return fmt.Errorf("graphql: %s", strings.Join(msgs, "; "))
		}
		if out != nil {
			if err := json.Unmarshal(r.Data, out); err != nil {
				return fmt.Errorf("decode data: %w", err)
			}
		}
		return nil
	}
	return fmt.Errorf("http: exceeded retry attempts")
}

// rateLimited returns true when the response indicates a GitHub primary or
// secondary rate-limit hit (429, or 403 with remaining=0).
func rateLimited(resp *http.Response) bool {
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if resp.StatusCode == http.StatusForbidden {
		if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining == "0" {
			return true
		}
	}
	return false
}

// rateLimitWait derives a sleep duration from response headers: Retry-After
// (secondary rate limits) takes precedence over X-RateLimit-Reset (primary).
// Returns a 60s floor if neither header is usable, capped at maxRateLimitSleep.
func rateLimitWait(resp *http.Response) time.Duration {
	if v := resp.Header.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
			return clampDuration(time.Duration(secs) * time.Second)
		}
	}
	if v := resp.Header.Get("X-RateLimit-Reset"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			wait := time.Until(time.Unix(ts, 0))
			if wait > 0 {
				return clampDuration(wait + time.Second) // +1s buffer
			}
		}
	}
	return 60 * time.Second
}

func clampDuration(d time.Duration) time.Duration {
	if d > maxRateLimitSleep {
		return maxRateLimitSleep
	}
	return d
}

// truncate shortens b to at most n bytes, backing up to the last valid UTF-8
// rune boundary so the result is always well-formed.
func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	cut := n
	for cut > 0 && !utf8.RuneStart(b[cut]) {
		cut--
	}
	return string(b[:cut]) + "…"
}

