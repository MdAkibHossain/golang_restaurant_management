package routes

import (
	controllers "restaurant-management/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/users", controllers.GetUsers())
	incomingRoutes.GET("/user/:user_id", controllers.GetUser())
	incomingRoutes.POST("/user/signup", controllers.Sugnup())
	incomingRoutes.POST("/user/login", controllers.Login())
}
