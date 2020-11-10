package jsonrpc1_0

import (
	"fmt"
	"github.com/patdz/jsonrpc"
	"github.com/patdz/jsonrpc/proto"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type CalcRequest struct {
	proto.AppRequestBase
	Params []Args `json:"params"`
}

func TestClient(t *testing.T) {
	go func() {
		StartServer()
	}()

	time.Sleep(time.Second)

	ob := &proto.DebugObserver{
		ProtoOutput: func(s string) {
			fmt.Printf("output: %v\n", s)
		},
		ProtoIncoming: func(s string) {
			fmt.Printf("incming: %v\n", s)
		},
	}

	client, err := jsonrpc.Dial(nil, ob, "tcp",
		fmt.Sprintf("127.0.0.1:%v", testServerPort))
	assert.Nil(t, err)

	req := &CalcRequest{
		AppRequestBase: proto.AppRequestBase{
			ServerMethod: "Calc.Compute",
		},
		Params: []Args{
			{
				A:  10,
				B:  20,
				Op: "+",
			},
		},
	}

	var result Reply
	resp := proto.AppResponse{
		Result: &result,
		Params: new(interface{}),
	}
	err2 := client.Call(req, &resp)
	assert.Nil(t, err2)
	assert.True(t, resp.Result != nil && result.Data == 30)

	req.Params[0].Op = "-"
	err2 = client.Call(req, &resp)
	assert.Nil(t, err2)
	assert.True(t, resp.Result != nil && result.Data == -10)

}
