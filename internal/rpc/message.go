package rpc

import "encoding/json"

type BaseRequest struct {
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Id      json.RawMessage `json:"id"`
}

type BaseResponse struct {
	Version string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	Id      json.RawMessage `json:"id"`
}

type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}
