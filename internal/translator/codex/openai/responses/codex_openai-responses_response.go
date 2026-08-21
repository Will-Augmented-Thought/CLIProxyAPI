package responses

import (
	"bytes"
	"context"

	translatorcommon "github.com/router-for-me/CLIProxyAPI/v7/internal/translator/common"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ConvertCodexResponseToOpenAIResponses converts OpenAI Chat Completions streaming chunks
// to OpenAI Responses SSE events (response.*).

func ConvertCodexResponseToOpenAIResponses(_ context.Context, modelName string, originalRequestRawJSON, requestRawJSON, rawJSON []byte, _ *any) [][]byte {
	hasDataPrefix := bytes.HasPrefix(rawJSON, []byte("data:"))
	if hasDataPrefix {
		rawJSON = bytes.TrimSpace(rawJSON[5:])
	}

	eventType := gjson.GetBytes(rawJSON, "type").String()
	if (eventType == "response.created" || eventType == "response.in_progress") && !gjson.GetBytes(rawJSON, "response.model").Exists() {
		requestModelName := translatorcommon.RequestModelName(originalRequestRawJSON, requestRawJSON)
		if requestModelName == "" {
			requestModelName = modelName
		}
		if requestModelName != "" {
			if updated, errSet := sjson.SetBytes(rawJSON, "response.model", requestModelName); errSet == nil {
				rawJSON = updated
			}
		}
	}
	if metadata := gjson.GetBytes(originalRequestRawJSON, "metadata"); metadata.Exists() && gjson.GetBytes(rawJSON, "response").IsObject() {
		if updated, errSet := sjson.SetRawBytes(rawJSON, "response.metadata", []byte(metadata.Raw)); errSet == nil {
			rawJSON = updated
		}
	}

	if hasDataPrefix {
		out := make([]byte, 0, len(rawJSON)+len("data: "))
		out = append(out, []byte("data: ")...)
		out = append(out, rawJSON...)
		return [][]byte{out}
	}
	return [][]byte{rawJSON}
}

// ConvertCodexResponseToOpenAIResponsesNonStream builds a single Responses JSON
// from a non-streaming OpenAI Chat Completions response.
func ConvertCodexResponseToOpenAIResponsesNonStream(_ context.Context, _ string, originalRequestRawJSON, _, rawJSON []byte, _ *any) []byte {
	rootResult := gjson.ParseBytes(rawJSON)
	// Verify this is a terminal response event.
	responseType := rootResult.Get("type").String()
	if responseType != "response.completed" && responseType != "response.incomplete" {
		return []byte{}
	}
	if metadata := gjson.GetBytes(originalRequestRawJSON, "metadata"); metadata.Exists() && rootResult.Get("response").IsObject() {
		if updated, errSet := sjson.SetRawBytes(rawJSON, "response.metadata", []byte(metadata.Raw)); errSet == nil {
			rawJSON = updated
			rootResult = gjson.ParseBytes(rawJSON)
		}
	}
	responseResult := rootResult.Get("response")
	return []byte(responseResult.Raw)
}
