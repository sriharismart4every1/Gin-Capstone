package models

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type ChargingStation struct {
	StationID              int    `json:"stationId" gorm:"type:int;primary_key"`
	EnergyOutput           string `json:"energyOutput"`
	Type                   string `json:"type"`
	VehicleBatteryCapacity string `json:"vehicleBatteryCapacity"`
	CurrentVehicleCharge   string `json:"currentVehicleCharge"`
	ChargingStartTime      string `json:"chargingStartTime"` //change 1
	AvailabiltyTime        string `json:"availabilityTime"`
}

var DB *gorm.DB

func DatabaseConnect() {
	db, _ := gorm.Open("mysql", "root:Srihari@129@tcp(127.0.0.1:3306)/")
	_ = db.Exec("CREATE DATABASE IF NOT EXISTS " + "chargestations")
	_ = db.Exec("USE " + "chargestations") //change 8
	db.AutoMigrate(&ChargingStation{})

	DB = db
	log.Println("DataBase connected successfully!!!")

}
