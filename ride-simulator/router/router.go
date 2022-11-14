package router

import (
	"net/http"
	"ride-simulator/database"
	"ride-simulator/entities"

	"github.com/gin-gonic/gin"
)

func InitializeRouter() {
	router := gin.Default()
	router.POST("api/v1/riders", postNewRider)
	router.POST("api/v1/drivers", postNewDriver)
	router.POST("api/v1/rides", postRideStartOrEndRequest)
	router.PUT("api/v1/drivers/:id/status", putDriverStatus)
	router.GET("api/v1/rides", GetRideStatusByDriverID)
	router.Run("localhost:8089")
}

func postNewRider(c *gin.Context) {
	var user entities.UserWithoutStatus
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
	}
	user, err := database.RegisterRider(user.Phone)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.IndentedJSON(http.StatusCreated, user)
}

func postNewDriver(c *gin.Context) {
	var user entities.UserWithoutStatus

	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
	}
	user, err := database.RegisterDriver(user.Phone)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.IndentedJSON(http.StatusCreated, user)
}

func putDriverStatus(c *gin.Context) {
	id := c.Param("id")
	var user entities.User
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
	}
	var onlineStatus bool
	if user.Status == "online" {
		onlineStatus = true
	} else {
		onlineStatus = false
	}
	user, err := database.UpdateDriversOnlineStatus(id, onlineStatus)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.IndentedJSON(http.StatusCreated, user)
}

func postRideStartOrEndRequest(c *gin.Context) {
	var ride entities.RideWithoutRiderPhone
	var err error

	if err = c.BindJSON(&ride); err != nil || (ride.RiderID == "" && ride.DriverID == "") {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if ride.DriverID != "" {
		ride, err = database.EndRide(ride.DriverID)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
	} else {
		ride, err = database.RequestRide(ride.RiderID)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
	}
	c.IndentedJSON(http.StatusCreated, ride)
}

func GetRideStatusByDriverID(c *gin.Context) {
	id := c.Query("driver_id")

	ride, err := database.GetCurrentRideStatusFromDriverID(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err})
		return
	} else {
		c.IndentedJSON(http.StatusOK, ride)
		return
	}
}
