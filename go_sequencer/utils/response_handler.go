package utils

import (
	"github.com/gin-gonic/gin"
)

func SuccessResponseHandler(c *gin.Context, status int, message string, data ...interface{}) {
	response := gin.H{
		"message": message,
	}
	if len(data) > 0 {
		response["data"] = data[0]
	}
	c.JSON(status, response)
}

func ErrorResponseHandler(c *gin.Context, status int, message string, err ...interface{}) {
	response := gin.H{
		"message": message,
	}
	if len(err) > 0 {
		response["error"] = err[0]
	}
	c.JSON(status, response)
}
