package xecho

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestError_Format(t *testing.T) {
	err := &Error{
		Status: http.StatusInternalServerError,
		Code:   "MY_ERROR",
		Detail: "My Error",
		Params: map[string]string{
			"Reason": "private reason",
		},
	}

	errString := err.Error()

	assert.Equal(t, "Code: MY_ERROR; Status: 500; Detail: My Error; Reason: private reason", errString)
}
