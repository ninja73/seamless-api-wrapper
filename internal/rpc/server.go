package rpc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"strings"
	"sync"
)

const Version = "2.0"

type server struct {
	method    map[string]HandlerFunc
	lock      sync.RWMutex
	transport Transport
	reqPool   *reqPool
}

func NewServer(transport Transport) *server {
	return &server{
		transport: transport,
		method:    make(map[string]HandlerFunc),
		reqPool:   new(reqPool),
	}
}

func (s *server) Run(ctx context.Context) error {
	if err := s.transport.Run(ctx, s); err != nil {
		return err
	}

	return nil
}

func (s *server) Register(name string, f HandlerFunc) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.method[name]; ok {
		return errors.New(fmt.Sprintf("method '%s' alredey exists", name))
	}

	s.method[strings.ToLower(name)] = f
	return nil
}

func (s *server) Resolve(ctx context.Context, w io.Writer, r io.Reader) {
	data := s.reqPool.get()
	defer s.reqPool.put(data)

	size, err := bufio.NewReader(r).Read(data)
	if err != nil {
		writeError(w, nil, err)
		return
	}

	data = bytes.TrimLeft(data[:size], " \t\r\n")
	var response []byte
	switch {
	case len(data) > 0 && data[0] == '[':
		var batch []*BaseRequest
		if err := json.Unmarshal(data, &batch); err != nil {
			writeError(w, nil, ParseError)
			return
		}

		if len(batch) == 0 {
			writeError(w, nil, InvalidReqError)
			return
		}

		result, err := s.batchReader(ctx, batch)
		if err != nil {
			writeError(w, nil, err)
			return
		}

		if len(result) == 0 {
			return
		}

		if response, err = json.Marshal(result); err != nil {
			writeError(w, nil, err)
			return
		}
	case len(data) > 0 && data[0] == '{':
		var req BaseRequest
		if err := json.Unmarshal(data, &req); err != nil {
			writeError(w, nil, ParseError)
			return
		}

		result, ok := s.singleReader(ctx, &req)
		if !ok {
			return
		}
		if response, err = json.Marshal(result); err != nil {
			writeError(w, nil, err)
			return
		}
	default:
		writeError(w, nil, InvalidReqError)
	}

	if _, err := w.Write(response); err != nil {
		log.Error(err)
	}
}

func (s *server) singleReader(ctx context.Context, req *BaseRequest) (*BaseResponse, bool) {
	if err := validateRequest(req); err != nil {
		return errorResponse(req.Id, err), true
	}

	h, err := s.getHandler(req)
	if err != nil {
		return errorResponse(req.Id, err), true
	}

	result, err := h(ctx, req.Params)
	if err != nil {
		return errorResponse(req.Id, &Error{
			Code:    ServerErrorCode,
			Message: err.Error(),
		}), true
	}

	if len(req.Id) == 0 {
		return nil, false
	}

	return successResponse(req.Id, result), true
}

func (s *server) batchReader(ctx context.Context, batch []*BaseRequest) ([]*BaseResponse, error) {
	var resultLock sync.Mutex
	var wg sync.WaitGroup

	result := make([]*BaseResponse, 0, len(batch))
	for i := range batch {
		baseReq := batch[i]

		wg.Add(1)
		go func(req *BaseRequest) {
			defer wg.Done()

			response, ok := s.singleReader(ctx, req)
			if !ok {
				return
			}

			resultLock.Lock()
			result = append(result, response)
			resultLock.Unlock()
		}(baseReq)
	}

	wg.Wait()

	return result, nil
}

func (s *server) getHandler(r *BaseRequest) (HandlerFunc, error) {
	s.lock.RLock()
	h, ok := s.method[strings.ToLower(r.Method)]
	s.lock.RUnlock()

	if !ok {
		return nil, MethodNotFoundError
	}

	return h, nil
}

func writeError(w io.Writer, id json.RawMessage, err error) {
	resp := errorResponse(id, err)
	data, err := json.Marshal(resp)
	if err != nil {
		log.Error(err)
	}

	if _, err := w.Write(data); err != nil {
		log.Error(err)
	}
}

func errorResponse(id json.RawMessage, err error) *BaseResponse {
	resp := &BaseResponse{
		Version: Version,
		Id:      id,
	}

	switch t := err.(type) {
	case *Error:
		resp.Error = t
	default:
		resp.Error = &Error{
			Code:    InternalErrorCode,
			Message: err.Error(),
		}
	}

	return resp
}

func successResponse(id json.RawMessage, resp json.RawMessage) *BaseResponse {
	baseResp := &BaseResponse{
		Version: Version,
		Id:      id,
		Result:  resp,
	}

	return baseResp
}

func validateRequest(req *BaseRequest) error {
	if req.JsonRPC != Version {
		return InvalidReqError
	}

	if len(req.Method) == 0 {
		return InvalidReqError
	}

	return nil
}
