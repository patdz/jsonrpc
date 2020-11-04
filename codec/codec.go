package codec

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/patdz/jsonrpc/helper"
	"github.com/patdz/jsonrpc/proto"
	"github.com/pkg/errors"
)

type clientCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer

	// JSON-RPC responses include the request id but not the request method.
	// Package rpc expects both.
	// We save the request method in pending when sending a request
	// and then look it up by request ID when filling out the rpc Response.
	mutex   sync.Mutex        // protects pending
	pending map[uint64]string // map request id to method name

	debug bool
}

// NewClientCodec returns a new rpc.ClientCodec using JSON-RPC on conn.
func NewClientCodec(conn io.ReadWriteCloser, debug bool) proto.ClientCodec {
	return &clientCodec{
		dec:     json.NewDecoder(conn),
		enc:     json.NewEncoder(conn),
		c:       conn,
		pending: make(map[uint64]string),
		debug:   debug,
	}
}

func (c *clientCodec) WriteRequest(r proto.AppRequest) error {
	c.mutex.Lock()
	c.pending[r.Seq()] = r.Method()
	c.mutex.Unlock()
	if c.debug {
		di, _ := json.Marshal(r)
		fmt.Printf("=> %v\n", string(di))
	}
	return c.enc.Encode(r)
}

type clientResponse struct {
	Id     uint64 `json:"id"`
	Method string `json:"method"`

	Raw interface{}
}

func (r *clientResponse) reset() {
	r.Id = 0
	r.Raw = nil
}

func (c *clientCodec) ReadResponseHeader(r *proto.Response) (err error) {
	var raw interface{}
	if err = c.dec.Decode(&raw); err != nil {
		return
	}

	if c.debug {
		di, _ := json.Marshal(&raw)
		fmt.Printf("<= %v\n", string(di))
	}

	mp, ok := raw.(map[string]interface{})
	if !ok {
		return errors.New("invalid response type")
	}

	r.ID, _ = helper.Interface2Uint64(mp["id"])
	r.Method, _ = helper.Interface2String(mp["method"])
	r.Error, _ = helper.Interface2JsonBytes(mp["error"])
	r.Params, _ = helper.Interface2JsonBytes(mp["params"])
	r.Result, _ = helper.Interface2JsonBytes(mp["result"])
	r.Resp = mp

	if r.ID == 0 {
		if r.Method == "" {
			r.CheckError = &proto.Error{
				Code:    proto.ErrorCodeInvalidResponse,
				Message: "method is nil",
			}
		}
		return nil
	}

	c.mutex.Lock()
	r.Method = c.pending[r.ID]
	delete(c.pending, r.ID)
	c.mutex.Unlock()

	if r.Error == nil && r.Result == nil {
		r.CheckError = &proto.Error{
			Code:    proto.ErrorCodeInvalidResponse,
			Message: "no result and error field",
		}
	}
	return nil
}

func (c *clientCodec) ReadResponseBody(resp *proto.Response, appResp *proto.AppResponse) *proto.Error {
	if resp == nil || appResp == nil {
		return nil
	}
	appResp.Resp = resp

	if appResp.Result != nil {
		if resp.Result != nil {
			err := json.Unmarshal(resp.Result, &appResp.Result)
			if err != nil {
				return &proto.Error{
					Code:    proto.ErrRespParseFailed,
					Message: fmt.Sprintf("prase result field failed: %v", err),
				}
			}
		}
	}

	if appResp.Params != nil {
		if resp.Params != nil {
			err := json.Unmarshal(resp.Params, &appResp.Params)
			if err != nil {
				return &proto.Error{
					Code:    proto.ErrRespParseFailed,
					Message: fmt.Sprintf("prase params field failed: %v", err),
				}
			}
		} else {
			appResp.Params = nil
		}
	}
	return nil
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}
