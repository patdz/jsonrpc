package jsonrpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/patdz/jsonrpc/proto"
	"github.com/sirupsen/logrus"
)

type MiningSubscribe struct {
	proto.AppRequestBase
	Params []interface{} `json:"params"`
}

func TestClient(t *testing.T) {
	t.SkipNow()

	logrus.SetLevel(logrus.DebugLevel)

	nc := &proto.NotificationChan{
		NewMessageChan: make(chan *proto.Response, 100),
		ErrorChan:      make(chan *proto.Error, 100),
	}

	ob := &proto.DebugObserver{
		ProtoOutput: func(s string) {
			fmt.Printf("output: %v\n", s)
		},
		ProtoIncoming: func(s string) {
			fmt.Printf("incming: %v\n", s)
		},
	}
	client, err := Dial(nc, ob, "tcp", "X.X.X.X:3333")
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
		AppRequestBase: proto.AppRequestBase{
			ServerMethod: "mining.subscribe",
		},
		Params: []interface{}{},
	}

	var result []interface{}
	resp := proto.AppResponse{
		Result: &result,
		Params: new(interface{}),
	}
	_ = client.Call(req, &resp)

	if resp.Result != nil {
		logrus.Info("Result not nil")
	}
	if resp.Params != nil {
		logrus.Info("Params not nil")
	}

	time.Sleep(10 * time.Second)
}
