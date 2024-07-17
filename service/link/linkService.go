package link

import (
	"errors"
	"net/http"
	"strconv"
	"tokamak-sybil-resistance/models"
	"tokamak-sybil-resistance/utils"
	"tokamak-sybil-resistance/utils/database"

	"github.com/gin-gonic/gin"
)

func CreateLink(c *gin.Context) {
	var link *models.Link
	err := c.ShouldBind(&link)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Error Creating an Link", err)
		return
	}
	res := database.Db.Create(link)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Error Creating an Link", err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, "Link Successfully Created", link)
}

func GetLinkById(c *gin.Context) {
	var link models.Link
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Invalid link ID", err.Error())
		return
	}
	res := database.Db.First(&link, idInt)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusNotFound, "Link not found", err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, "Link fetched successfully", link)
}

func GetAllLinks(c *gin.Context) {
	var links []models.Link
	res := database.Db.Find(&links)
	if res.Error != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Error Fetching Links", errors.New("links not found"))
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, "Fetched all the Available Links", links)
}

func UpdateLink(c *gin.Context) {
	var link models.Link
	id := c.Param("id")
	err := c.ShouldBind(&link)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Error Updating an Link", err)
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Invalid link ID", err.Error())
		return
	}
	var updateLink models.Account
	res := database.Db.Model(&updateLink).Where("id = ?", idInt).Updates(link)

	if res.Error != nil || res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Link not Updated", err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, "Link Updated Successfully", link)
}

func DeleteLink(c *gin.Context) {
	var link models.Link
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Invalid link ID", err.Error())
		return
	}
	res := database.Db.First(&link, idInt)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, "Link not found", err)
		return
	}
	database.Db.Delete(&link)
	utils.SuccessResponseHandler(c, http.StatusOK, "Link Deleted Successfully")
}
