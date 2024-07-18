package utils

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CheckError(res *gorm.DB, c *gin.Context, httpStatus int, responseMessage string, err error) bool {
	if err != nil || res != nil && res.Error != nil || res.RowsAffected == 0 {
		ErrorResponseHandler(c, httpStatus, responseMessage, err)
		return true
	}
	return false
}
