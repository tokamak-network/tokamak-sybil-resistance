package account

import (
	"net/http"
	"strconv"
	"tokamak-sybil-resistance/database"
	"tokamak-sybil-resistance/database/statedb"
	"tokamak-sybil-resistance/lib/response_messages"
	"tokamak-sybil-resistance/models"
	"tokamak-sybil-resistance/utils"

	"github.com/gin-gonic/gin"
)

type Account struct {
	stateDB *statedb.StateDB
}

func NewAccount(stateDB *statedb.StateDB) Account {
	return Account{
		stateDB: stateDB,
	}
}

func (a *Account) CreateAccount(c *gin.Context) {
	var account models.Account
	err := c.ShouldBind(&account)
	if utils.CheckError(nil, c, http.StatusBadRequest, response_messages.ErrorCreatingAccount, err) {
		return
	}
	// a.stateDB.Put()
	// res := database.Db.Create(&account)
	// if utils.CheckError(res, c, http.StatusBadRequest, response_messages.ErrorCreatingAccount, err) {
	// 	return
	// }
	// utils.SuccessResponseHandler(c, http.StatusOK, response_messages.AccountCreationSuccess, account)
}

func GetAccountById(c *gin.Context) {
	var account models.Account
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if utils.CheckError(nil, c, http.StatusBadRequest, response_messages.InvalidAccountId, err) {
		return
	}
	res := database.Db.First(&account, idInt)
	if utils.CheckError(res, c, http.StatusNotFound, response_messages.AccountNotFound, err) {
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.FetchAccountByIdSuccess, account)
}

func GetAllAccount(c *gin.Context) {
	var accounts []models.Account
	res := database.Db.Find(&accounts)
	if utils.CheckError(res, c, http.StatusBadRequest, response_messages.ErrorFetchingAccount, nil) {
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.FetchAllAccounts, accounts)
}

func UpdateAccount(c *gin.Context) {
	var account models.Account
	id := c.Param("id")
	err := c.ShouldBind(&account)
	if utils.CheckError(nil, c, http.StatusBadRequest, response_messages.ErrorUpdatingAccount, err) {
		return
	}
	idInt, err := strconv.Atoi(id)
	if utils.CheckError(nil, c, http.StatusBadRequest, response_messages.InvalidAccountId, err) {
		return
	}
	var updateAccount models.Account
	res := database.Db.Model(&updateAccount).Where("id = ?", idInt).Updates(&account)
	if utils.CheckError(res, c, http.StatusBadRequest, response_messages.AccountNotUpdated, err) {
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.AccountUpdateSuccess, account)
}

func DeleteAccount(c *gin.Context) {
	var account models.Account
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if utils.CheckError(nil, c, http.StatusBadRequest, response_messages.InvalidAccountId, err) {
		return
	}
	res := database.Db.First(&account, idInt)
	if utils.CheckError(res, c, http.StatusBadRequest, response_messages.AccountNotFound, err) {
		return
	}
	database.Db.Delete(&account)
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.AccountDeletedSuccessfully)
}
