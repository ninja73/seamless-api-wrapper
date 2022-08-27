package rpc

import (
	"context"
	"encoding/json"
)

type HandlerFunc func(context.Context, json.RawMessage) (json.RawMessage, error)

func HandlerWithPointer[RQ any, RS any](handler func(context.Context, *RQ) (RS, error)) HandlerFunc {
	return func(ctx context.Context, in json.RawMessage) (json.RawMessage, error) {
		req := new(RQ)
		if err := json.Unmarshal(in, req); err != nil {
			return nil, InvalidParamsError
		}

		resp, err := handler(ctx, req)
		if err != nil {
			return nil, &Error{
				Code:    ServerErrorCode,
				Message: err.Error(),
			}
		}
		return json.Marshal(resp)
	}
}

func HandlerWithoutParams[RS any](handler func(context.Context) (RS, error)) HandlerFunc {
	return func(ctx context.Context, in json.RawMessage) (json.RawMessage, error) {
		resp, err := handler(ctx)
		if err != nil {
			return nil, &Error{
				Code:    ServerErrorCode,
				Message: err.Error(),
			}
		}
		return json.Marshal(resp)
	}
}

func Handler[RQ any, RS any](handler func(context.Context, RQ) (RS, error)) HandlerFunc {
	return func(ctx context.Context, in json.RawMessage) (json.RawMessage, error) {
		var req RQ
		if err := json.Unmarshal(in, &req); err != nil {
			return nil, InvalidParamsError
		}
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, &Error{
				Code:    ServerErrorCode,
				Message: err.Error(),
			}
		}
		return json.Marshal(resp)
	}
}
