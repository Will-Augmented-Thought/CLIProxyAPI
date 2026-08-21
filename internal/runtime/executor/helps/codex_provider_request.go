package helps

import "github.com/tidwall/sjson"

// StripCodexUnsupportedMetadata removes OpenAI Responses metadata immediately
// before transport because the ChatGPT Codex provider endpoint rejects it.
func StripCodexUnsupportedMetadata(body []byte) []byte {
	body, _ = sjson.DeleteBytes(body, "metadata")
	return body
}
