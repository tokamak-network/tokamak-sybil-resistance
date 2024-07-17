package account

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"tokamak-sybil-resistance/models"
	"tokamak-sybil-resistance/utils"
	"tokamak-sybil-resistance/utils/database"

	"github.com/gin-gonic/gin"
)

func CreateAccount(c *gin.Context) {
	var account *models.Account
	fmt.Println(account)
	err := c.ShouldBind(&account)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Error Creating an Account", err)
		return
	}
	res := database.Db.Create(account)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Error Creating an Account", err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, "Account Successfully Created", account)
}

func GetAccountById(c *gin.Context) {
	var account models.Account
	id := c.Param("id")
	fmt.Println(id)
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Invalid account ID", err.Error())
		return
	}
	res := database.Db.First(&account, idInt)
	fmt.Println(res)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusNotFound, "Account not found", err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, "Account fetched successfully", account)
}

func GetAllAccount(c *gin.Context) {
	fmt.Printf("we are here")
	var accounts []models.Account
	res := database.Db.Find(&accounts)
	fmt.Println(res)
	if res.Error != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Error Fetching Accounts", errors.New("authors not found"))
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, "Fetched all the Available Account Data", accounts)
}

func UpdateAccount(c *gin.Context) {
	var account models.Account
	fmt.Println(account)
	id := c.Param("id")
	err := c.ShouldBind(&account)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Error Updating an Account", err)
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Invalid account ID", err.Error())
		return
	}
	var updateAccount models.Account
	fmt.Println(updateAccount)
	res := database.Db.Model(&updateAccount).Where("id = ?", idInt).Updates(account)

	if res.Error != nil || res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Account not Updated", err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, "Link Updated Successfully", account)
}

func DeleteAccount(c *gin.Context) {
	var account models.Account
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Invalid account ID", err.Error())
		return
	}
	res := database.Db.First(&account, idInt)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Account not found", err)
		return
	}
	database.Db.Delete(&account)
	utils.SuccessResponseHandler(c, http.StatusOK, "Account Deleted Successfully")
}
