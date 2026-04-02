package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ohler55/ojg/jp"
	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"
)

type HTTPRunner struct {
	client *http.Client
	state  *state.Store
}

func NewHTTPRunner(s *state.Store) *HTTPRunner {
	return &HTTPRunner{
		client: &http.Client{},
		state:  s,
	}
}

func (r *HTTPRunner) Run(base string, step workflow.Step) (any, error) {
	if step.Request == nil {
		return nil, fmt.Errorf("http step request must not be nil")
	}

	vars := r.state.All()
	url := template.Render(step.Request.URL, vars)
	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("%s%s", base, url)
	}

	var body io.Reader
	if step.Request.Body != nil {
		b, _ := json.Marshal(step.Request.Body)
		bodyContent := template.Render(string(b), vars)
		body = bytes.NewReader([]byte(bodyContent))
	}

	req, err := http.NewRequest(step.Request.Method, url, body)
	if err != nil {
		return nil, err
	}
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

	result := map[string]any{
		"status": resp.StatusCode,
		"data":   decoded,
		"header": resp.Header,
	}

	for key, path := range step.Extract {
		if p, err := jp.ParseString(path); err == nil {
			values := p.Get(result)
			if len(values) > 0 {
				r.state.Set(key, values[0])
			}
		}
	}

	return result, nil
}
