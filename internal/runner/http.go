package runner

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/replay/replay/internal/reporter"
	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"
)

type HTTPRunner struct {
	client   *http.Client
	state    *state.Store
	reporter *reporter.Reporter
}

func NewHTTPRunner(s *state.Store, rep *reporter.Reporter) *HTTPRunner {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &HTTPRunner{
		client:   &http.Client{Transport: tr},
		state:    s,
		reporter: rep,
	}
}

func (r *HTTPRunner) Run(config workflow.HTTPConfig, step workflow.Step) (any, error) {
	if step.Request == nil {
		return nil, fmt.Errorf("http step request must not be nil")
	}

	vars := r.state.All()
	baseURL := template.Render(config.BaseURL, vars)
	url := template.Render(step.Request.URL, vars)

	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("%s%s", baseURL, url)
	}

	var body io.Reader
	var bodyRaw any
	if step.Request.Body != nil {
		bodyRaw = step.Request.Body
		// If it's a string, we might need to interpolate and it might be raw JSON
		if s, ok := bodyRaw.(string); ok {
			bodyContent := template.Render(s, vars)
			body = bytes.NewReader([]byte(bodyContent))
			bodyRaw = bodyContent
		} else {
			// If it's a map/slice, marshal to JSON, then interpolate
			b, _ := json.Marshal(bodyRaw)
			bodyContent := template.Render(string(b), vars)
			body = bytes.NewReader([]byte(bodyContent))
			// Decode back for cleaner debug log
			json.Unmarshal([]byte(bodyContent), &bodyRaw)
		}
	}

	if config.Debug {
		r.reporter.Debug("HTTP REQUEST", map[string]any{
			"method":  step.Request.Method,
			"url":     url,
			"headers": step.Request.Headers,
			"body":    bodyRaw,
		})
	}

	req, err := http.NewRequest(step.Request.Method, url, body)
	if err != nil {
		return nil, err
	}

	// Apply global headers first
	for k, v := range config.Headers {
		req.Header.Set(k, template.Render(v, vars))
	}

	// Apply step-specific headers (overrides global)
	for k, v := range step.Request.Headers {
		req.Header.Set(k, template.Render(v, vars))
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var decoded any
	json.Unmarshal(respBody, &decoded)

	if config.Debug {
		r.reporter.Debug("HTTP RESPONSE", map[string]any{
			"status":  resp.StatusCode,
			"headers": resp.Header,
			"body":    decoded,
		})
	}

	result := map[string]any{
		"status": resp.StatusCode,
		"data":   decoded,
		"header": resp.Header,
	}

	for key, path := range step.Extract {
		// Interpolate path in case it contains variables (dynamic extraction)
		path = template.Render(path, vars)

		// Handle "res." prefix for consistency with assertions
		cleanPath := path
		if strings.HasPrefix(path, "res.") {
			cleanPath = strings.Replace(path, "res.", "$.", 1)
		} else if !strings.HasPrefix(path, "$") {
			cleanPath = "$." + path
		}

		expr, err := ParseJSONPath(cleanPath, step.Name, key)
		if err != nil {
			return result, err
		}

		values := expr.Get(result)
		if len(values) > 0 {
			r.state.Set(key, values[0])
		}
	}

	return result, nil
}
