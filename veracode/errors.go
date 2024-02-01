package veracode

import "fmt"

type UserError struct {
	Method   string
	Url      string
	Message  string `json:"message"`
	HttpCode int    `json:"http_code"`
	// HttpStatus string `json:"http_status"`
	UserId       string
	EmailAddress string
	err          error
}

func (e *UserError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return fmt.Sprintf("%s %s(%d): %s", e.Method, e.Url, e.HttpCode, e.Message)
}
