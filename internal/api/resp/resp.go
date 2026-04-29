package resp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/pkg/errors"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

func Error(c *gin.Context, httpCode int, err *errors.Error) {
	c.JSON(httpCode, Response{Code: err.Code, Message: err.Message})
}

func BadRequest(c *gin.Context, err *errors.Error) { Error(c, http.StatusBadRequest, err) }
func Unauthorized(c *gin.Context, err *errors.Error) { Error(c, http.StatusUnauthorized, err) }
func NotFound(c *gin.Context, err *errors.Error)     { Error(c, http.StatusNotFound, err) }
func InternalError(c *gin.Context, err *errors.Error) { Error(c, http.StatusInternalServerError, err) }

func ParseUintParam(c *gin.Context, name string) (uint, error) {
	var id uint
	if _, err := fmt.Sscanf(c.Param(name), "%d", &id); err != nil {
		BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, "invalid "+name))
		return 0, err
	}
	return id, nil
}
