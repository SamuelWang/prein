package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct{}

func (c *Controller) GetItems(ctx *gin.Context) {
	// Logic to retrieve items
	ctx.JSON(http.StatusOK, gin.H{"message": "GetItems called"})
}

func (c *Controller) CreateItem(ctx *gin.Context) {
	// Logic to create an item
	ctx.JSON(http.StatusCreated, gin.H{"message": "CreateItem called"})
}