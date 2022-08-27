package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

type TestTransport struct {
}

func (t *TestTransport) Run(_ context.Context, _ Resolver) error {
	return nil
}

func subtract(_ context.Context, data []int) (int, error) {
	if len(data) < 2 {
		return 0, nil
	}

	return data[0] - data[1], nil
}

func TestResolve(t *testing.T) {
	var srv = NewServer(&TestTransport{})
	_ = srv.Register("subtract", Handler(subtract))

	jsonObj := `{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}`
	expected := `{"jsonrpc":"2.0","result":19,"id":1}`

	out := bytes.NewBuffer([]byte{})
	in := bytes.NewReader([]byte(jsonObj))

	ctx := context.Background()
	srv.Resolve(ctx, out, in)

	if out.String() != expected {
		t.Errorf("got %q, expected %q", out.String(), expected)
	}

	jsonObj = `{"jsonrpc": "2.0", "method": "subtract", "params": [23, 42], "id": 2}`
	expected = `{"jsonrpc":"2.0","result":-19,"id":2}`

	out.Reset()
	in.Reset([]byte(jsonObj))
	srv.Resolve(ctx, out, in)

	if out.String() != expected {
		t.Errorf("got %q, expected %q", out.String(), expected)
	}
}

type SubtractData struct {
	Subtrahend int `json:"subtrahend"`
	Minuend    int `json:"minuend"`
}

func subtract2(_ context.Context, data *SubtractData) (int, error) {
	return data.Minuend - data.Subtrahend, nil
}

func TestResolveNamedParams(t *testing.T) {
	var srv = NewServer(&TestTransport{})
	_ = srv.Register("subtract", HandlerWithPointer(subtract2))

	jsonObj := `{"jsonrpc": "2.0", "method": "subtract", "params": {"subtrahend": 23, "minuend": 42}, "id": 3}`
	expected := `{"jsonrpc":"2.0","result":19,"id":3}`

	out := bytes.NewBuffer([]byte{})
	in := bytes.NewReader([]byte(jsonObj))

	ctx := context.Background()
	srv.Resolve(ctx, out, in)

	if out.String() != expected {
		t.Errorf("got %q, expected %q", out.String(), expected)
	}

	jsonObj = `{"jsonrpc": "2.0", "method": "subtract", "params": {"minuend": 42, "subtrahend": 23}, "id": 4}`
	expected = `{"jsonrpc":"2.0","result":19,"id":4}`

	out.Reset()
	in.Reset([]byte(jsonObj))
	srv.Resolve(ctx, out, in)

	if out.String() != expected {
		t.Errorf("got %q, expected %q", out.String(), expected)
	}
}

func TestResolveError(t *testing.T) {
	var srv = NewServer(&TestTransport{})
	_ = srv.Register("subtract", Handler(subtract))

	jsonObj := `{"jsonrpc": "2.0", "method": "foobar, "params": "bar", "baz]`
	expected := `{"jsonrpc":"2.0","error":{"code":-32700,"message":"parse error"},"id":null}`

	out := bytes.NewBuffer([]byte{})
	in := bytes.NewReader([]byte(jsonObj))

	ctx := context.Background()
	srv.Resolve(ctx, out, in)

	if out.String() != expected {
		t.Errorf("got %q, expected %q", out.String(), expected)
	}

	jsonObj = `[
  {"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},
  {"jsonrpc": "2.0", "method"
]`
	expected = `{"jsonrpc":"2.0","error":{"code":-32700,"message":"parse error"},"id":null}`

	out.Reset()
	in.Reset([]byte(jsonObj))
	srv.Resolve(ctx, out, in)

	if out.String() != expected {
		t.Errorf("got %q, expected %q", out.String(), expected)
	}

	jsonObj = `[]`
	expected = `{"jsonrpc":"2.0","error":{"code":-32600,"message":"invalid request"},"id":null}`

	out.Reset()
	in.Reset([]byte(jsonObj))
	srv.Resolve(ctx, out, in)

	if out.String() != expected {
		t.Errorf("got %q, expected %q", out.String(), expected)
	}
}

func TestResolveBatch(t *testing.T) {
	var srv = NewServer(&TestTransport{})
	_ = srv.Register("subtract", Handler(subtract))

	jsonObj := `[
        {"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},
        {"jsonrpc": "2.0", "method": "subtract", "params": [42,23], "id": "2"},
        {"foo": "boo"}
    ]`

	expected := `[{"jsonrpc":"2.0","error":{"code":-32601,"message":"The method does not exist."},"id":"1"},{"jsonrpc":"2.0","result":19,"id":"2"},{"jsonrpc":"2.0","error":{"code":-32600,"message":"invalid request"},"id":null}]`

	var responses []BaseResponse

	_ = json.Unmarshal([]byte(expected), &responses)
	expectedResponseMap := make(map[string]struct{}, len(responses))
	for _, resp := range responses {
		expectedResponseMap[string(resp.Id)] = struct{}{}
	}

	out := bytes.NewBuffer([]byte{})
	in := bytes.NewReader([]byte(jsonObj))

	ctx := context.Background()
	srv.Resolve(ctx, out, in)

	responses = responses[:0]
	_ = json.Unmarshal(out.Bytes(), &responses)
	responseMap := make(map[string]struct{}, len(responses))
	for _, resp := range responses {
		responseMap[string(resp.Id)] = struct{}{}
	}

	for id := range responseMap {
		delete(expectedResponseMap, id)
	}

	if len(expectedResponseMap) > 0 {
		t.Errorf("got %q, expected %q", out.String(), expected)
	}

	jsonObj = `[
        {"jsonrpc": "2.0", "method": "subtract", "params": [1,2,4]},
        {"jsonrpc": "2.0", "method": "subtract", "params": [7]}
    ]`
	expected = ``

	out.Reset()
	in.Reset([]byte(jsonObj))
	srv.Resolve(ctx, out, in)

	if out.String() != expected {
		t.Errorf("got %q, expected %q", out.String(), expected)
	}
}
