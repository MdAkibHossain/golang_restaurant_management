package main

import (
	"os"
	middleware "restaurant-management/middleware"
	routes "restaurant-management/routes"

	"github.com/gin-gonic/gin"
)

//var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.Authentication())

	routes.UserRoutes(router)
	routes.FoodRoutes(router)
	routes.InvoiceRoutes(router)
	routes.MenuRoutes(router)
	routes.OrderItemRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)

	router.Run(": " + port)

}
