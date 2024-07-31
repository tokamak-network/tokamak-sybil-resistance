package main

import (
	"tokamak-sybil-resistance/database"
	"tokamak-sybil-resistance/database/statedb"
	"tokamak-sybil-resistance/routes"
	"tokamak-sybil-resistance/service/account"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()
	stateDB := statedb.InitNewStateDB()
	account.NewAccount(stateDB)
	router := gin.Default()
	routes.Account(router)
	routes.Link(router)
	router.Run("localhost:8080")
}
