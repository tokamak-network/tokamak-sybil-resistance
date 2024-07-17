package routes

import (
	"tokamak-sybil-resistance/service/account"

	"github.com/gin-gonic/gin"
)

func Account(router *gin.Engine) {
	router.POST("/account", account.CreateAccount)
	router.GET("/account/:id", account.GetAccountById)
	router.GET("/account", account.GetAllAccount)
	router.PUT("/account/:id", account.UpdateAccount)
	router.DELETE("/account/:id", account.DeleteAccount)
}
