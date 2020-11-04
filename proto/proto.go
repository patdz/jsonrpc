package proto

const (
	ErrorCodeInvalidResponse = -42700
	ErrorShutdown            = -42701
	ErrRespParseFailed       = -42702
)

// Response is a header written before every RPC return. It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
type Response struct {
	ID     uint64
	Method string
	Error  []byte
	Params []byte
	Result []byte
	Resp   map[string]interface{}

	CheckError *Error
}

type Error struct {
	Code    int
	Message string
}

// A ClientCodec implements writing of RPC requests and
// reading of RPC responses for the client side of an RPC session.
// The client calls WriteRequest to write a request to the connection
// and calls ReadResponseHeader and ReadResponseBody in pairs
// to read responses. The client calls Close when finished with the
// connection. ReadResponseBody may be called with a nil
// argument to force the body of the response to be read and then
// discarded.
// See NewClient's comment for information about concurrent access.
type ClientCodec interface {
	WriteRequest(AppRequest) error
	ReadResponseHeader(*Response) error
	ReadResponseBody(*Response, *AppResponse) *Error

	Close() error
}

type NotificationChan struct {
	NewMessageChan chan *Response
	ErrorChan      chan *Error
}

// AppRequest: must include json[id, method, params]
type AppRequest interface {
	SetSeq(uint64)
	Seq() uint64
	Method() string
}

type AppResponse struct {
	Resp   *Response
	Result interface{}
	Params interface{}
	Error  interface{}
}
