package utils


type HyperliquidError interface {
	error
}

type ClientError struct {
	StatusCode   int
	ErrorCode    string
	ErrorMessage string
	Header       map[string][]string
	ErrorData    interface{}
}

// Error implements the error interface for ClientError.
func (e *ClientError) Error() string {
	return e.ErrorMessage
}

type ServerError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface for ServerError.
func (e *ServerError) Error() string {
	return e.Message
}