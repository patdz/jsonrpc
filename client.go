package jsonrpc

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"

	"github.com/patdz/jsonrpc/codec"
	"github.com/patdz/jsonrpc/proto"
)

var debugLog = false

var ErrShutdown = errors.New("connection is shut down")

// Call represents an active RPC.
type Call struct {
	Request  proto.AppRequest // The argument to the function (*struct).
	Response proto.Response
	Reply    *proto.AppResponse // The reply from the function (*struct).
	Error    *proto.Error
	Done     chan *Call // Receives *Call when Go is complete.
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	codec proto.ClientCodec

	reqMutex sync.Mutex // protects following

	mutex    sync.Mutex // protects following
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // user has called Close
	shutdown bool // server has told us to stop

	newMessageChan chan *proto.Response
	errorChan      chan *proto.Error
}

func (client *Client) send(call *Call) {
	client.reqMutex.Lock()
	defer client.reqMutex.Unlock()

	// Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		client.mutex.Unlock()
		call.Error = &proto.Error{Code: proto.ErrorShutdown, Message: "client shutdown"}
		call.done()
		return
	}
	seq := client.seq
	client.seq++
	client.pending[seq] = call
	client.mutex.Unlock()

	call.Request.SetSeq(seq)

	// Encode and send the request.
	err := client.codec.WriteRequest(call.Request)
	if err != nil {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = &proto.Error{Code: proto.ErrorShutdown, Message: err.Error()}
			call.done()
		}
	}
}

func (client *Client) input() {
	var err error
	for {
		response := &proto.Response{}
		err = client.codec.ReadResponseHeader(response)
		if err != nil {
			if client.errorChan != nil {
				client.errorChan <- &proto.Error{
					Code:    proto.ErrorCodeInvalidResponse,
					Message: err.Error(),
				}
			}
			break
		}
		seq := response.ID
		if seq == 0 {
			if client.newMessageChan != nil {
				client.newMessageChan <- response
			}
			continue
		}

		var call *Call

		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()

		switch {
		case call == nil:
			if client.errorChan != nil {
				client.errorChan <- &proto.Error{
					Code:    proto.ErrorCodeInvalidResponse,
					Message: "caller is null",
				}
			}
		case response.CheckError != nil:
			call.Error = response.CheckError
			call.done()
		default:
			call.Error = client.codec.ReadResponseBody(response, call.Reply)
			call.done()
		}
	}
	// Terminate pending calls.
	client.reqMutex.Lock()
	client.mutex.Lock()
	client.shutdown = true
	closing := client.closing
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = &proto.Error{
			Code:    proto.ErrorShutdown,
			Message: "shutdown",
		}
		call.done()
	}
	client.mutex.Unlock()
	client.reqMutex.Unlock()
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here. It is the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
		if debugLog {
			log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
		}
	}
}

// NewClientWithCodec is like NewClient but uses the specified
// codec to encode requests and decode responses.
func NewClientWithCodec(notifyChan *proto.NotificationChan, codec proto.ClientCodec) *Client {
	var newMessageChan chan *proto.Response
	var errorChan chan *proto.Error
	if notifyChan != nil {
		newMessageChan = notifyChan.NewMessageChan
		errorChan = notifyChan.ErrorChan
	}
	client := &Client{
		codec:          codec,
		pending:        make(map[uint64]*Call),
		newMessageChan: newMessageChan,
		errorChan:      errorChan,
		seq:            1,
	}
	go client.input()
	return client
}

// Close calls the underlying codec's Close method. If the connection is already
// shutting down, ErrShutdown is returned.
func (client *Client) Close() error {
	client.mutex.Lock()
	if client.closing {
		client.mutex.Unlock()
		return ErrShutdown
	}
	client.closing = true
	client.mutex.Unlock()
	return client.codec.Close()
}

// Go invokes the function asynchronously. It returns the Call structure representing
// the invocation. The done channel will signal when the call is complete by returning
// the same Call object. If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (client *Client) Go(req proto.AppRequest, reply *proto.AppResponse, done chan *Call) *Call {
	call := new(Call)
	call.Request = req
	call.Reply = reply
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel. If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	client.send(call)
	return call
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(req proto.AppRequest, reply *proto.AppResponse) *proto.Error {
	call := <-client.Go(req, reply, make(chan *Call, 1)).Done
	return call.Error
}

// Dial connects to a JSON-RPC server at the specified network address.
func Dial(notifyChan *proto.NotificationChan, debug bool, network, address string) (*Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClientWithCodec(notifyChan, codec.NewClientCodec(conn, debug)), err
}
