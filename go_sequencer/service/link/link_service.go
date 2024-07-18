package link

import (
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
	if utils.CheckError(nil, c, http.StatusBadRequest, response_messages.ErrorCreatingLink, err) {
		return
	}
	res := database.Db.Create(&link)
	if utils.CheckError(res, c, http.StatusBadRequest, response_messages.ErrorCreatingLink, err) {
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.LinkCreationSuccesss, link)
}

func GetLinkById(c *gin.Context) {
	var link models.Link
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if utils.CheckError(nil, c, http.StatusBadRequest, response_messages.InvalidLinkId, err) {
		return
	}
	res := database.Db.First(&link, idInt)
	if utils.CheckError(res, c, http.StatusNotFound, response_messages.LinkNotFound, err) {
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.LinkFetchedByIdSuccess, link)
}

func GetAllLinks(c *gin.Context) {
	var links []models.Link
	res := database.Db.Find(&links)
	if utils.CheckError(res, c, http.StatusBadRequest, response_messages.ErrorFetchingLink, nil) {
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.FetchedAllLinks, links)
}

func UpdateLink(c *gin.Context) {
	var link models.Link
	id := c.Param("id")
	err := c.ShouldBind(&link)
	if utils.CheckError(nil, c, http.StatusBadRequest, response_messages.ErrorUpdatingLink, err) {
		return
	}
	idInt, err := strconv.Atoi(id)
	if utils.CheckError(nil, c, http.StatusBadRequest, response_messages.InvalidLinkId, err) {
		return
	}
	var updateLink models.Link
	res := database.Db.Model(&updateLink).Where("id = ?", idInt).Updates(&link)
	if utils.CheckError(res, c, http.StatusBadRequest, response_messages.ErrorLinkNotUpdated, err) {
		return
	}
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.LinkUpdatedSuccess, link)
}

func DeleteLink(c *gin.Context) {
	var link models.Link
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if utils.CheckError(nil, c, http.StatusBadRequest, response_messages.InvalidLinkId, err) {
		return
	}
	res := database.Db.First(&link, idInt)
	if utils.CheckError(res, c, http.StatusBadRequest, response_messages.LinkNotFound, err) {
		return
	}
	database.Db.Delete(&link)
	utils.SuccessResponseHandler(c, http.StatusOK, response_messages.LinkDeleteSuccess)
}
