package routes

import (
	"backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
    controller := controllers.Controller{}

    router.GET("/items", controller.GetItems)
    router.POST("/items", controller.CreateItem)
}