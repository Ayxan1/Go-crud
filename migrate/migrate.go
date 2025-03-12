package main

import (
	"example.com/gocrud/initializers"
	"example.com/gocrud/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectDB()
}

func main() {
	initializers.DB.AutoMigrate(&models.Post{})
}
