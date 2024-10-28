package json

import (
	jsoniter "github.com/json-iterator/go"
)

// RawMessage is a type alias for jsoniter.RawMessage to behave like json.RawMessage
type RawMessage = jsoniter.RawMessage

// json is the standard configuration for jsoniter, designed to behave like encoding/json
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Functions that operate similarly to encoding/json
var (
	Marshal       = json.Marshal
	Unmarshal     = json.Unmarshal
	MarshalIndent = json.MarshalIndent
	NewDecoder    = json.NewDecoder
	NewEncoder    = json.NewEncoder
)
