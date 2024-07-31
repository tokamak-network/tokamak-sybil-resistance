package main

import (
	"tokamak-sybil-resistance/database/statedb"

	"github.com/gin-gonic/gin"
)

func main() {
	statedb.InitNewStateDB()
	router := gin.Default()
	router.Run("localhost:8080")
}
