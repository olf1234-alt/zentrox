package zentrox

// HTTPError is the canonical error payload returned by the framework.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

func (e HTTPError) Error() string { return e.Message }

// NewHTTPError constructs a new HTTPError as a Go error.
func NewHTTPError(code int, message string, detail ...any) error {
	var d any
	if len(detail) > 0 {
		d = detail[0]
	}
	return HTTPError{Code: code, Message: message, Detail: d}
}
