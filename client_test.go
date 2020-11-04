package jsonrpc

import (
	"testing"
	"time"

	"github.com/patdz/jsonrpc/proto"
	"github.com/sirupsen/logrus"
)

type AppRequestBase struct {
	SeqNumber1   uint64        `json:"id"`
	ServerMethod string        `json:"method"`
	Params       []interface{} `json:"params"`
}

func (r *AppRequestBase) SetSeq(seq uint64) {
	r.SeqNumber1 = seq
}
func (r *AppRequestBase) Seq() uint64 {
	return r.SeqNumber1
}

func (r *AppRequestBase) Method() string {
	return r.ServerMethod
}

type MiningSubscribe struct {
	AppRequestBase
}

func TestClient(t *testing.T) {
	t.SkipNow()

	logrus.SetLevel(logrus.DebugLevel)

	nc := &proto.NotificationChan{
		NewMessageChan: make(chan *proto.Response, 100),
		ErrorChan:      make(chan *proto.Error, 100),
	}
	client, err := Dial(nc, true, "tcp", "X.X.X.X:3333")
	if err != nil {
		panic(err)
	}

	go func() {
		for n := range nc.NewMessageChan {
			logrus.Debugf("--- %v\n", n.Method)
		}
	}()

	go func() {
		for n := range nc.ErrorChan {
			logrus.Debugf("--- %v\n", n)
		}
	}()

	req := &MiningSubscribe{
		AppRequestBase: AppRequestBase{
			ServerMethod: "mining.subscribe",
			Params:       []interface{}{},
		},
	}

	resp := proto.AppResponse{
		Result: new(interface{}),
		Params: new(interface{}),
	}
	_ = client.Call(req, &resp)

	if *(resp.Result) != nil {
		logrus.Info("Result not nil")
	}
	if *(resp.Params) != nil {
		logrus.Info("Params not nil")
	}

	time.Sleep(10 * time.Second)
}
