package models

import (
	"encoding/json"
	"io/ioutil"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB

type University struct {
	Rank int    `json:"ranking" gorm:"primary_key"`
	Name string `json:"title" gorm:"type:varchar(100)"`
	City string `json:"location" gorm:"type:varchar(100)"`
}

var institutions []University

func DBConnect() {
	db, _ := gorm.Open("mysql", "root:Srihari@129@tcp(127.0.0.1:3306)/")
	db.Exec("CREATE DATABASE IF NOT EXISTS " + "universities")
	db.Exec("USE " + "universities")
	db.AutoMigrate(&University{})
	log.Println("Connected to database successfully")
	file, err := ioutil.ReadFile("json/universities_ranking.json")

	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(file), &institutions)

	for _, val := range institutions {
		db.Where(University{Rank: val.Rank}).Assign(University{Rank: val.Rank, Name: val.Name, City: val.City}).FirstOrCreate(&University{})
	}
	log.Println("Successfully unmarshalled data from json and stored in database!!")
	DB = db

}
