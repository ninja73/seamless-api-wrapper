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
			return nil, err
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
			return nil, err
		}
		return json.Marshal(resp)
	}
}
