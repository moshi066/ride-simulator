package main

import (
	"ride-simulator/database"
	"ride-simulator/router"
)

func main() {
	database.InitializeDatabase()
	router.InitializeRouter()
}
