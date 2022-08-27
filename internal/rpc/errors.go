package rpc

type ErrorCode = int

const (
	ParseErrorCode     = -32700
	InvalidRequestCode = -32600
	MethodNotFoundCode = -32601
	InvalidParamsCode  = -32602
	InternalErrorCode  = -32603
	ServerErrorCode    = -32000
)

var (
	ParseError          = &Error{Code: ParseErrorCode, Message: "parse error"}
	InvalidReqError     = &Error{Code: InvalidRequestCode, Message: "invalid request"}
	MethodNotFoundError = &Error{Code: MethodNotFoundCode, Message: "The method does not exist."}
	InvalidParamsError  = &Error{Code: InvalidParamsCode, Message: "invalid method parameters"}
)
