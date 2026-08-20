package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/thinking"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
)

const (
	codexSparkModel          = "gpt-5.3-codex-spark"
	codexSparkExhaustedModel = "gpt-5.4-mini"
)

type codexSparkQuotaExhaustionError struct {
	cause error
}

func (e *codexSparkQuotaExhaustionError) Error() string {
	if e == nil || e.cause == nil {
		return ""
	}
	return e.cause.Error()
}

func (e *codexSparkQuotaExhaustionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// codexSparkQuotaRoute is the single product-defined capacity transition. It
// is deliberately not a generic fallback mechanism: only a fully exhausted
// Codex Spark credential pool may execute one Mini request.
type codexSparkQuotaRoute struct {
	enabled   bool
	providers []string
}

func (r codexSparkQuotaRoute) Next(err error, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Request, cliproxyexecutor.Options, bool) {
	var exhausted *codexSparkQuotaExhaustionError
	if !r.enabled || len(r.providers) != 1 || r.providers[0] != "codex" ||
		strings.TrimSpace(thinking.ParseSuffix(req.Model).ModelName) != codexSparkModel ||
		!errors.As(err, &exhausted) || exhausted == nil {
		return cliproxyexecutor.Request{}, cliproxyexecutor.Options{}, false
	}

	miniReq := req
	miniReq.Model = codexSparkExhaustedModel
	miniOpts := ensureRequestedModelMetadata(opts, req.Model)
	if len(miniOpts.Metadata) == 0 {
		return miniReq, miniOpts, true
	}

	metadata := make(map[string]any, len(miniOpts.Metadata))
	for key, value := range miniOpts.Metadata {
		if key != cliproxyexecutor.AuthSelectionModelMetadataKey {
			metadata[key] = value
		}
	}
	miniOpts.Metadata = metadata
	return miniReq, miniOpts, true
}

func (r codexSparkQuotaRoute) ExecutedModelHeaders(headers http.Header) http.Header {
	if headers == nil {
		headers = make(http.Header)
	}
	headers.Set("OpenAI-Model", codexSparkExhaustedModel)
	return headers
}
