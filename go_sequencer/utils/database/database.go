package database

import (
	"fmt"
	"log"
	"tokamak-sybil-resistance/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB
var err error

/*
*
TODO - Create a config package and move these db configuration keys there
*
*/
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "root"
	dbname   = "postgres"
)

var modelsToMigrate = []interface{}{
	&models.Account{},
	&models.Link{},
}

func InitDB() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failure to connect db")
		log.Fatal(err)
	}
	fmt.Println("Database connected!")

	err = Db.AutoMigrate(modelsToMigrate...)
	if err != nil {
		log.Fatalf("Failure to migrate db")
		log.Fatal(err)
	}
	fmt.Println("Database migrated!")
}
