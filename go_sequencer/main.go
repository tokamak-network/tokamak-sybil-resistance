package main

import (
	"tokamak-sybil-resistance/routes"
	"tokamak-sybil-resistance/utils/database"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()
	router := gin.Default()
	routes.Account(router)
	routes.Link(router)
	router.Run("localhost:8080")
}
