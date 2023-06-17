package main

import (
	"gin_capstone/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

type ChargingStationDemo struct {
	StationID              int    `json:"stationId" gorm:"type:int;primary_key"`
	EnergyOutput           string `json:"energyOutput"`
	Type                   string `json:"type"`
	VehicleBatteryCapacity string `json:"vehicleBatteryCapacity"`
	CurrentVehicleCharge   string `json:"currentVehicleCharge"`
	ChargingStartTime      string `json:"chargingStartTime"` // change2
	AvailabiltyTime        string `json:"availabilityTime"`
}

type SubStruct struct {
	StationID    int    `json:"stationId" gorm:"type:int;primary_key"`
	EnergyOutput string `json:"energyOutput"`
	Type         string `json:"type"`
}

type SubStruct1 struct {
	StationID              int    `json:"stationId" gorm:"type:int;primary_key"`
	VehicleBatteryCapacity string `json:"vehicleBatteryCapacity"`
	CurrentVehicleCharge   string `json:"currentVehicleCharge"`
	ChargingStartTime      string `json:"chargingStartTime"` // change3
}

type SubStruct2 struct {
	StationID       int    `json:"stationId" gorm:"type:int;primary_key"`
	EnergyOutput    string `json:"energyOutput"`
	Type            string `json:"type"`
	AvailabiltyTime string `json:"availabilityTime"`
}

// Add Charging Stations
func AddChargingStation(ctx *gin.Context) {
	var input ChargingStationDemo
	//var display SubStruct
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": "Not able to post data"})
		log.Println("Charging Station Not Added!!")
	}
	details := models.ChargingStation{EnergyOutput: input.EnergyOutput, Type: input.Type}
	models.DB.Create(&details)
	log.Printf("Charging Station %d Added Successfully", details.StationID)
	dp := SubStruct{details.StationID, details.EnergyOutput, details.Type}
	ctx.JSON(http.StatusCreated, gin.H{"data": dp})
}

// Start Charging
func StartCharging(ctx *gin.Context) {
	var station models.ChargingStation
	var input ChargingStationDemo

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": "Not able to post data"})
		log.Println("Not able to bind station")
	}
	if err := models.DB.Where(input.StationID).First(&station).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": "Station Id doesn't exists!!"})
		log.Println("Station Id doesn't exists!!")
		return
	}

	models.DB.Model(&station).Updates(input)
	log.Printf("Successfully started charging station %d", input.StationID)
	dp := SubStruct1{station.StationID, station.VehicleBatteryCapacity, station.CurrentVehicleCharge, station.ChargingStartTime}
	ctx.JSON(http.StatusCreated, gin.H{"data": dp})

}

// Available Charging Stations
func AvailableChargingStations(ctx *gin.Context) {
	var stations []models.ChargingStation
	var substations []SubStruct
	//var val SubStruct
	//var check time.Time // change4
	//change5
	if err := models.DB.Where("charging_start_time = ?", "").Find(&stations).Error; err == nil {
		for _, sub := range stations {
			val := SubStruct{sub.StationID, sub.EnergyOutput, sub.Type}
			substations = append(substations, val)
		}
		log.Println("Successfully fetched available charging stations")
		ctx.JSON(http.StatusOK, gin.H{"data": substations})
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"Error": "No Available Charging Stations!!"})
	log.Println("No available charging stations")

}

// Occupied charging stations
func OccupiedChargingStations(ctx *gin.Context) {
	var stations []models.ChargingStation
	//var stationsDemo []models.ChargingStation
	var substations []SubStruct2
	//var check time.Time //change6
	//change7
	if err := models.DB.Not("charging_start_time = ?", "").Find(&stations).Error; err == nil {
		for _, val := range stations {
			val.AvailabiltyTime = calculateAvailabilityTime(val)
			inp := SubStruct2{val.StationID, val.EnergyOutput, val.Type, val.AvailabiltyTime}
			substations = append(substations, inp)
			models.DB.Save(&val)
		}
		log.Println("Successfully fetched occupied charging stations")
		ctx.JSON(http.StatusOK, gin.H{"data": substations})
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"Error": "No Occupied Charging Stations!!"})
	log.Println("No Occupied charging stations")

}

func calculateAvailabilityTime(ch models.ChargingStation) string {

	vbc, _ := strconv.Atoi(ch.VehicleBatteryCapacity[:len(ch.VehicleBatteryCapacity)-3])
	cvc, _ := strconv.Atoi(ch.CurrentVehicleCharge[:len(ch.CurrentVehicleCharge)-3])
	eo, _ := strconv.Atoi(ch.EnergyOutput[:len(ch.EnergyOutput)-3])

	timeInHours := int64(vbc-cvc) / int64(eo)

	val, _ := time.Parse(time.RFC3339, ch.ChargingStartTime)

	result := val.Add(time.Duration(timeInHours) * time.Hour)

	ch.AvailabiltyTime = result.Format("2006-01-02T15:04:05Z")
	log.Println("Successfullly calculated availability time!!")
	return ch.AvailabiltyTime

}

var localcache = cache.New(5*time.Minute, 10*time.Minute)

var newcache *cache.Cache = cache.New(1*time.Minute, 3*time.Minute)

func SetCache(stationId int, station models.ChargingStation) bool {
	localcache.Set(strconv.Itoa(stationId), station, cache.NoExpiration)
	return true
}

func getCache(stationId string) (interface{}, bool, string) {
	var source string
	data, found := localcache.Get(stationId)
	if found {
		source = "Cache"
	}
	return data, found, source

}

// GetStationById
func GetStationById(ctx *gin.Context) {
	var station models.ChargingStation
	var source string
	data, err, source := getCache(ctx.Param("stationId"))
	if !err {
		if err := models.DB.Where("station_id=?", ctx.Param("stationId")).First(&station).Error; err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"data": "Record Not Found"})
			log.Printf("Station Id %s doesn't exists", ctx.Param("stationId"))
		} else {
			SetCache(station.StationID, station)
			log.Printf("Set data in cache successfull")
			source = "database"
			ctx.JSON(http.StatusOK, gin.H{"data": station, "source": source})
			log.Println("Successfully fetched data from database")
		}
	} else {
		val := data.(models.ChargingStation)
		ctx.JSON(http.StatusOK, gin.H{"data": val, "source": source})
		log.Println("Successfully fetched data from cache")
	}
}

// GetAllStations
func GetAllStations(ctx *gin.Context) {
	var stations []models.ChargingStation
	cache_key := "charge"
	if result, found := newcache.Get(cache_key); found {
		log.Println("Successfully fetced data from cache!!")
		ctx.JSON(http.StatusOK, gin.H{"data": result, "source": "cache"})
		return
	}
	models.DB.Find(&stations)
	newcache.Set(cache_key, stations, cache.DefaultExpiration)
	ctx.JSON(http.StatusOK, gin.H{"data": stations, "source": "database"})
	log.Println("Successfully fetced data from database!!")
}

// Logging
func logging() {
	logfile, err := os.Create("app.log")
	if err != nil {
		log.Fatal(err)

	}

	log.SetOutput(logfile)
	log.Println("Application Started.....!!!")
	log.Println("Logging Started!!!")
}

func main() {
	logging()
	router := gin.Default()
	models.DatabaseConnect()
	router.GET("/api/occupiedchargestations", OccupiedChargingStations)
	router.GET("/api/availablechargestations", AvailableChargingStations)
	router.POST("/api/startcharging", StartCharging)
	router.POST("/api/addchargestation", AddChargingStation)
	// Rest End Points Using cache
	router.GET("/api/getallstations", GetAllStations)
	router.GET("/api/getstationbyid/:stationId", GetStationById)
	router.Run("127.0.0.1:5001")
}
