package resolvereview

import "encoding/json"

func mustJSON(v any) json.RawMessage {
	raw, _ := json.Marshal(v)
	return raw
}
