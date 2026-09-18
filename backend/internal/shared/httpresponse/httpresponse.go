package httpresponse

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Data: data})
}

func Failure(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, Envelope{Error: &ErrorBody{Code: code, Message: message}})
}
