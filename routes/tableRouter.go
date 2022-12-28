package routes

import (
	controllers "restaurant-management/controllers"

	"github.com/gin-gonic/gin"
)

func TableRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/table", controllers.GetTables())
	incomingRoutes.GET("/table/:table_id", controllers.GetTable())
	incomingRoutes.POST("/table", controllers.CreateTable())
	incomingRoutes.PATCH("/table/:table_id", controllers.UpdateTable())
}
