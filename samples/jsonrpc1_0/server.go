package jsonrpc1_0

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

const (
	testServerPort = 9999
)

type Calc struct{}

type Args struct {
	A  float64 `json:"a"`
	B  float64 `json:"b"`
	Op string  `json:"op"`
}

type Reply struct {
	Msg  string  `json:"msg"`
	Data float64 `json:"data"`
}

func (c *Calc) Compute(args Args, reply *Reply) error {
	var (
		msg string = "ok"
	)

	switch args.Op {
	case "+":
		reply.Data = args.A + args.B
	case "-":
		reply.Data = args.A - args.B
	case "*":
		reply.Data = args.A * args.B
	case "/":
		if args.B == 0 {
			msg = "in divide op, B can't be zero"
		} else {
			reply.Data = args.A / args.B
		}
	default:
		msg = fmt.Sprintf("unsupported op:%s", args.Op)
	}
	reply.Msg = msg

	if reply.Msg == "ok" {
		return nil
	}
	return fmt.Errorf(msg)
}

func StartServer() {
	_ = rpc.Register(new(Calc))

	l, err := net.Listen("tcp", fmt.Sprintf(":%v", testServerPort))
	if err != nil {
		log.Fatalln("listen error:", err)
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		go jsonrpc.ServeConn(conn)
	}
}
