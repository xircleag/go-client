package common

import (
	"fmt"
)

type RequestError struct {
	StatusCode int         `json:"-"`
	Code       int         `json:"code"`
	ID         string      `json:"id"`
	Message    string      `json:"message"`
	URL        string      `json:"url"`
	Data       interface{} `json:"data"`
}

func (e RequestError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}
