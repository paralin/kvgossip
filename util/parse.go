package util

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
)

// Fast o(1) check for JSON, then uses proto unmarshal.
func HeuristicParse(data []byte, target proto.Message) error {
	// Detect JSON
	if len(data) > 2 &&
		(data[0] == byte('{') ||
			(data[0] == 0 && data[1] == byte('{'))) {
		if json.Unmarshal(data, target) == nil {
			return nil
		}
	}

	return proto.Unmarshal(data, target)
}
