package routes

import (
	"tokamak-sybil-resistance/service/link"

	"github.com/gin-gonic/gin"
)

func Link(router *gin.Engine) {
	router.POST("/link", link.CreateLink)
	router.GET("/link/:id", link.GetLinkById)
	router.GET("/link", link.GetAllLinks)
	router.PUT("/link/:id", link.UpdateLink)
	router.DELETE("/link/:id", link.DeleteLink)
}
