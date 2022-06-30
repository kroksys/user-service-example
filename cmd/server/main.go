package main

import (
	"log"

	"github.com/kroksys/user-service-example/pkg/db"
)

func main() {
	// Connect to database
	err := db.Connect("user:userpw@tcp(localhost:3306)/users?parseTime=true")
	if err != nil {
		log.Fatalf("Error connecting to database: %s\n", err.Error())
	}

	// Migrate user model to database
	err = db.Migrate()
	if err != nil {
		log.Fatalf("Error migrating user to database: %s\n", err.Error())
	}

}
