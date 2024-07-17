package account

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"tokamak-sybil-resistance/lib/response_messages"
	"tokamak-sybil-resistance/models"
	"tokamak-sybil-resistance/utils"
	"tokamak-sybil-resistance/utils/database"

	"github.com/gin-gonic/gin"
)

func CreateAccount(c *gin.Context) {
	var account models.Account
	fmt.Println(c)
	err := c.ShouldBind(&account)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.ErrorCreatingAccount, err)
		return
	}
	fmt.Println(account)
	res := database.Db.Create(&account)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.ErrorCreatingAccount, err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.AccountCreationSuccess, account)
}

func GetAccountById(c *gin.Context) {
	var account models.Account
	id := c.Param("id")
	fmt.Println(id)
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.InvalidAccountId, err.Error())
		return
	}
	res := database.Db.First(&account, idInt)
	fmt.Println(res)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusNotFound, response_messages.AccountNotFound, err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.FetchAccountByIdSuccess, account)
}

func GetAllAccount(c *gin.Context) {
	fmt.Printf("we are here")
	var accounts []models.Account
	res := database.Db.Find(&accounts)
	fmt.Println(res)
	if res.Error != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.ErrorFetchingAccount, errors.New("authors not found"))
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.FetchAllAccounts, accounts)
}

func UpdateAccount(c *gin.Context) {
	var account models.Account
	fmt.Println(account)
	id := c.Param("id")
	err := c.ShouldBind(&account)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.ErrorUpdatingAccount, err)
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.InvalidAccountId, err.Error())
		return
	}
	var updateAccount models.Account
	fmt.Println(updateAccount)
	res := database.Db.Model(&updateAccount).Where("id = ?", idInt).Updates(&account)

	if res.Error != nil || res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.AccountNotUpdated, err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.AccountUpdateSuccess, account)
}

func DeleteAccount(c *gin.Context) {
	var account models.Account
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.InvalidAccountId, err.Error())
		return
	}
	res := database.Db.First(&account, idInt)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.AccountNotFound, err)
		return
	}
	database.Db.Delete(&account)
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.AccountDeletedSuccessfully)
}
