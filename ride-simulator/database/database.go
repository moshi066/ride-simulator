package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"ride-simulator/entities"
)

var db *sql.DB

func InitializeDatabase() {
	dbName := "ride_simulator"
	var err error
	db, err = sql.Open("mysql", os.Getenv("DBUSER")+":"+os.Getenv("DBPASS")+"@tcp(127.0.0.1:3306)/")
	if err != nil {
		log.Fatal(err)
	}

	rows := db.QueryRow("SHOW DATABASES LIKE '" + dbName + "'")
	if err := rows.Scan(&dbName); err != nil {
		if err == sql.ErrNoRows {
			InitDatabaseForTheFirstTime(dbName)
			PopulateDatabaseWithMockData()
		} else {
			log.Fatal(err)
		}
	}

	db, err = sql.Open("mysql", os.Getenv("DBUSER")+":"+os.Getenv("DBPASS")+"@tcp(127.0.0.1:3306)/"+dbName)
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
}

func RegisterRider(phone string) (entities.UserWithoutStatus, error) {
	var user entities.UserWithoutStatus
	_, err := db.Exec("INSERT INTO users (id, phone, is_driver, status) VALUES (UUID(), ?, ?, ?)", phone, false, false)
	if err != nil {
		return user, err
	}

	row := db.QueryRow("SELECT id FROM users WHERE phone = ? AND is_driver = FALSE", phone)
	if err := row.Scan(&user.ID); err != nil {
		return user, err
	}
	user.Phone = phone
	user.Type = "rider"

	return user, nil
}

func RegisterDriver(phone string) (entities.UserWithoutStatus, error) {
	var user entities.UserWithoutStatus
	_, err := db.Exec("INSERT INTO users (id, phone, is_driver, status) VALUES (UUID(), ?, ?, ?)", phone, true, false)
	if err != nil {
		return user, err
	}
	row := db.QueryRow("SELECT id FROM users WHERE phone = ? AND is_driver = TRUE", phone)
	if err := row.Scan(&user.ID); err != nil {
		return user, err
	}
	user.Phone = phone
	user.Type = "driver"

	return user, nil
}

func GetCurrentRideStatusFromDriverID(driverID string) (entities.Ride, error) {
	var ride entities.Ride
	var status bool
	row := db.QueryRow("SELECT * FROM rides WHERE driver_id = ? ORDER BY id desc LIMIT 1", driverID)
	if err := row.Scan(&ride.ID, &ride.DriverID, &ride.RiderID, &status); err != nil {
		if err == sql.ErrNoRows {
			return ride, fmt.Errorf("no ride found with driver_id: %s", driverID)
		}
		return ride, err
	}
	if status {
		ride.Status = "end"
	} else {
		ride.Status = "start"
	}

	row = db.QueryRow("SELECT phone FROM users where id = ?", ride.RiderID)
	if err := row.Scan(&ride.RiderPhone); err != nil {
		return ride, err
	}
	return ride, nil
}

func UpdateDriversOnlineStatus(id string, onlineStatus bool) (entities.User, error) {
	var user entities.User

	res, err := db.Exec("UPDATE users SET status = ? WHERE id = ?", onlineStatus, id)
	if err != nil {
		return user, err
	} else {
		if rowsAffected, err := res.RowsAffected(); rowsAffected == 0 || err != nil {
			return user, fmt.Errorf("data wasn't updated due to an error: %v", err)
		}
	}

	row := db.QueryRow("SELECT phone FROM users WHERE id = ?", id)
	if err := row.Scan(&user.Phone); err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("no user found with id: %s", id)
		}
		return user, fmt.Errorf("user by id %s: %v", id, err)
	}
	if onlineStatus {
		user.Status = "online"
	} else {
		user.Status = "offline"
	}
	user.Type = "driver"
	user.ID = id
	return user, nil
}

func RequestRide(riderID string) (entities.RideWithoutRiderPhone, error) {
	var ride entities.RideWithoutRiderPhone
	var status bool
	var availableDriver string

	row := db.QueryRow("SELECT is_driver FROM users WHERE id = ?", riderID)
	if err := row.Scan(&status); err != nil {
		return ride, err
	}
	if status {
		return ride, fmt.Errorf("Driver can't request a ride!")
	}

	row = db.QueryRow("SELECT rider_id FROM rides WHERE rider_id = ? AND is_ride_ended = FALSE", riderID)
	if err := row.Scan(&riderID); err != nil {
		if err != sql.ErrNoRows {
			return ride, err
		}
	} else {
		return ride, fmt.Errorf("Can't request a new trip while a trip is running!")
	}

	row = db.QueryRow("SELECT users.id FROM users WHERE users.is_driver = TRUE AND users.id NOT IN (SELECT DISTINCT(rides.driver_id) FROM rides WHERE rides.is_ride_ended = FALSE) LIMIT 1")
	if err := row.Scan(&availableDriver); err != nil {
		if err == sql.ErrNoRows {
			return ride, fmt.Errorf("No driver available now!")
		}
		return ride, err
	}

	fmt.Println(availableDriver)

	_, err := db.Exec("INSERT INTO rides (id, driver_id, rider_id, is_ride_ended) VALUES (UUID(), ?, ?, ?)", availableDriver, riderID, false)
	if err != nil {
		return ride, err
	}

	ride.DriverID = availableDriver
	ride.RiderID = riderID
	ride.Status = "start"

	row = db.QueryRow("SELECT id FROM rides WHERE rider_id = ? AND driver_id = ? AND is_ride_ended = false", ride.RiderID, ride.DriverID)
	if err := row.Scan(&ride.ID); err != nil {
		return ride, err
	}

	return ride, nil
}

func EndRide(driverID string) (entities.RideWithoutRiderPhone, error) {
	var ride entities.RideWithoutRiderPhone
	var status bool
	var rideID string

	row := db.QueryRow("SELECT is_driver FROM users WHERE id = ?", driverID)
	if err := row.Scan(&status); err != nil {
		return ride, err
	}
	if !status {
		return ride, fmt.Errorf("Rider can't end a ride!")
	}
	row = db.QueryRow("SELECT id FROM rides WHERE driver_id = ? AND is_ride_ended = FALSE", driverID)
	if err := row.Scan(&rideID); err != nil {
		if err == sql.ErrNoRows {
			return ride, fmt.Errorf("No running ride found with driver_id: %s", driverID)
		} else {
			return ride, err
		}
	}
	_, err := db.Exec("UPDATE rides SET is_ride_ended = true WHERE id = ?", rideID)
	if err != nil {
		return ride, err
	}

	row = db.QueryRow("SELECT * FROM rides WHERE id = ?", rideID)
	if err := row.Scan(&ride.ID, &ride.DriverID, &ride.RiderID, &status); err != nil {
		return ride, err
	}
	ride.Status = "end"

	return ride, nil
}

func InitDatabaseForTheFirstTime(dbName string) {
	_, err := db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("USE " + dbName)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE Users (id VARCHAR(36) NOT NULL, phone VARCHAR(15) NOT NULL, is_driver boolean NOT NULL, status boolean NOT NULL, PRIMARY KEY (`id`), CONSTRAINT phone_is_driver_unique UNIQUE(phone, is_driver));")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE Rides (id VARCHAR(36) NOT NULL, driver_id VARCHAR(36) NOT NULL, rider_id VARCHAR(36) NOT NULL, is_ride_ended boolean NOT NULL, PRIMARY KEY (`id`));")
	if err != nil {
		log.Fatal(err)
	}
}

func PopulateDatabaseWithMockData() {
	_, err := db.Exec("INSERT INTO USERS (id, phone, is_driver, status) VALUES (UUID(), ?, ?, ?)", "01701021612", false, false)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO USERS (id, phone, is_driver, status) VALUES (UUID(), ?, ?, ?)", "01701021613", false, true)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO USERS (id, phone, is_driver, status) VALUES (UUID(), ?, ?, ?)", "01701021614", true, false)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO USERS (id, phone, is_driver, status) VALUES (UUID(), ?, ?, ?)", "01701021615", true, true)
	if err != nil {
		log.Fatal(err)
	}
}
