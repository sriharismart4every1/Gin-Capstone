package main

import (
	"log"
	"main/models"
	"net/http"
	"os"
	"strconv"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// Creating cache instance with default expiration
var localCache = cache.New(2*time.Minute, 5*time.Minute)

// Method to Set cache
func SetCache(rank string, univ models.University) bool {
	rankInt, err := strconv.Atoi(rank)
	if err != nil {
		panic(err)
	}
	if rankInt <= 20 { // No expiration for ranks below 21
		localCache.Set(rank, univ, cache.NoExpiration)
		//log.Printf("Successfully stored in cache for rank %d", rankInt)
	} else { // default expiration for ranks greater than 20
		localCache.Set(rank, univ, cache.DefaultExpiration)
		//log.Printf("Successfully stored in cache for rank %d", rankInt)
	}
	return true

}

// Method to get cache using key
func GetCache(rank string) (interface{}, bool, string) {
	var source string
	data, found := localCache.Get(rank) // will get data stored in cache based on the rank
	if found {
		source = "cache"
		//log.Printf("Successfully fetched data from cache for rank %s", rank)
	}
	return data, found, source

}

// Get All the Universities
func getAllUniversities(c *gin.Context) {
	var institutions []models.University
	cache_key := "data"                     // setting key for cache
	val, found := localCache.Get(cache_key) // using the cache key we are getting data from cache
	if found {                              // returns data from cache if found
		c.JSON(http.StatusOK, gin.H{"source": "cache", "data": val})
		log.Println("Successfully fetched all universities from cache")
	} else { // otherwise returns from database and stores in the cache using key
		models.DB.Find(&institutions)
		localCache.Set(cache_key, institutions, cache.DefaultExpiration)
		log.Println("Data is not present in cache so it is fetched from database and stored in cache")
		c.JSON(http.StatusOK, gin.H{"source": "database", "data": institutions})

	}

}

// Get University By Rank
func getUniversity(c *gin.Context) {
	var institution models.University
	rankUpd, _ := strconv.Atoi(c.Param("rank"))
	data, err, source := GetCache(c.Param("rank"))
	if !err {
		if err := models.DB.Where(models.University{Rank: rankUpd}).First(&institution).Error; err != nil {
			log.Printf("No record found for rank %d", rankUpd)
			c.JSON(http.StatusBadRequest, gin.H{"Error": "No record found!!"})
			return
		}
		SetCache(c.Param("rank"), institution)
		source = "database"
		log.Printf("Successfully fetched university from database for rank %d and stored in cache", rankUpd)
		c.JSON(http.StatusOK, gin.H{"data": institution, "source": source})

	} else {
		institution = data.(models.University)
		log.Printf("Successfully fetched university from cache for rank %d", rankUpd)
		c.JSON(http.StatusOK, gin.H{"data": institution, "source": source})
	}

}

// Add University
func addUniversity(c *gin.Context) {
	var university models.University
	if err := c.ShouldBindJSON(&university); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"Error": "Not able to Add University"})
		log.Println("Failed to add university!!")
		return
	}
	models.DB.Create(&university)
	log.Println("Successfully added university!!")
	c.JSON(http.StatusCreated, gin.H{"data": university})
}

// Update University data by Rank
func updateUniversity(c *gin.Context) {
	var institution models.University
	var input models.University
	rankUpd, _ := strconv.Atoi(c.Param("rank"))
	if err := models.DB.Where(models.University{Rank: rankUpd}).First(&institution).Error; err != nil {
		log.Printf("University Not Found for rank %d", rankUpd)
		c.JSON(http.StatusNotFound, gin.H{"Error": "University Not Found!!!"})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Unable to update university for rank %d", rankUpd)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"Error": "Not able to Update data"})
		return
	}
	log.Printf("Successfully updated university for rank %d", rankUpd)
	models.DB.Model(&institution).Updates(&input)
	c.JSON(http.StatusOK, gin.H{"data": institution})

}

// Delete University By Rank
func deleteUniversity(c *gin.Context) {
	var university models.University
	rankUpd, err := strconv.Atoi(c.Param("rank"))
	if err != nil {
		panic(err)
	}

	if err := models.DB.First(&university, rankUpd).Error; err != nil {
		log.Printf("University Not Found for rank %d", rankUpd)
		c.JSON(http.StatusNotFound, gin.H{"Error": "University Not Found!!"})
		return
	}
	if err := models.DB.Delete(&models.University{}, c.Param("rank")).Error; err != nil {
		log.Printf("Failed to delete university with rank %d", rankUpd)
		c.JSON(500, gin.H{"Error": "Failed to delete university"})
		return
	}
	log.Printf("Successfully deleted university with rank %d", rankUpd)
	c.JSON(http.StatusOK, gin.H{"Status": "Deleted University succcessfully!!"})

}

func LogFunc() {
	logfile, err := os.Create("app.log")

	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logfile)
	log.Println("Application Started!!!")
	log.Println("Logging Started!!!")
}

func main() {
	LogFunc()
	models.DBConnect()
	router := gin.Default()
	router.POST("/api/adduniversity", addUniversity)
	router.PUT("/api/updateuniversity/:rank", updateUniversity)
	router.DELETE("/api/deluniversity/:rank", deleteUniversity)
	router.GET("/api/getuniversity/:rank", getUniversity)
	router.GET("/api/getuniversities", getAllUniversities)
	router.Run(":5000")
}
