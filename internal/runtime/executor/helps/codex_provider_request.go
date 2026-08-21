package helps

import "github.com/tidwall/sjson"

// StripCodexUnsupportedContext removes OpenAI Responses context fields
// immediately before transport because the ChatGPT Codex provider endpoint
// rejects them.
func StripCodexUnsupportedContext(body []byte) []byte {
	body, _ = sjson.DeleteBytes(body, "conversation")
	body, _ = sjson.DeleteBytes(body, "metadata")
	return body
}
