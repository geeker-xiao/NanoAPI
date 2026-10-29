package httpclient

import (
	"encoding/json"
	"strings"
)

func prettyJSON(body []byte, truncated bool) string {
	var value any
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber()
	if err := dec.Decode(&value); err != nil {
		return string(body)
	}
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return string(body)
	}
	if truncated {
		return string(out) + "\n\n... preview truncated ..."
	}
	return string(out)
}

func isPrintable(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if r < 32 {
			return false
		}
	}
	return true
}
