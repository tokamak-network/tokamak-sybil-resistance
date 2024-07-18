package link

import (
	"errors"
	"net/http"
	"strconv"
	"tokamak-sybil-resistance/lib/response_messages"
	"tokamak-sybil-resistance/models"
	"tokamak-sybil-resistance/utils"
	"tokamak-sybil-resistance/utils/database"

	"github.com/gin-gonic/gin"
)

func CreateLink(c *gin.Context) {
	var link models.Link
	err := c.ShouldBind(&link)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.ErrorCreatingLink, err)
		return
	}
	res := database.Db.Create(&link)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.ErrorCreatingLink, err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.LinkCreationSuccesss, link)
}

func GetLinkById(c *gin.Context) {
	var link models.Link
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.InvalidLinkId, err.Error())
		return
	}
	res := database.Db.First(&link, idInt)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusNotFound, response_messages.LinkNotFound, err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.LinkFetchedByIdSuccess, link)
}

func GetAllLinks(c *gin.Context) {
	var links []models.Link
	res := database.Db.Find(&links)
	if res.Error != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.ErrorFetchingLink, errors.New("links not found"))
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.FetchedAllLinks, links)
}

func UpdateLink(c *gin.Context) {
	var link models.Link
	id := c.Param("id")
	err := c.ShouldBind(&link)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.ErrorUpdatingLink, err)
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.InvalidLinkId, err.Error())
		return
	}
	var updateLink models.Link
	res := database.Db.Model(&updateLink).Where("id = ?", idInt).Updates(&link)

	if res.Error != nil || res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.ErrorLinkNotUpdated, err)
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.LinkUpdatedSuccess, link)
}

func DeleteLink(c *gin.Context) {
	var link models.Link
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.InvalidLinkId, err.Error())
		return
	}
	res := database.Db.First(&link, idInt)
	if res.RowsAffected == 0 {
		utils.ErrorResponseHandler(c, http.StatusBadRequest, response_messages.LinkNotFound, err)
		return
	}
	database.Db.Delete(&link)
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.LinkDeleteSuccess)
}
