package main

import (
	"flag"
	"fmt"
	"os"

	authmodels "github.com/HappyNaCl/new-debyut/internal/auth/models"
	"github.com/HappyNaCl/new-debyut/internal/common/database"
	eventmodels "github.com/HappyNaCl/new-debyut/internal/event/models"
	"github.com/HappyNaCl/new-debyut/internal/seeders"
	usersmodels "github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/joho/godotenv"
)

func main() {
	seed := flag.Bool("seed", false, "seed the database with initial data")
	flag.Parse()

	if err := godotenv.Load("../.env.local"); err != nil {
		panic(fmt.Sprintf("Failed to load .env.local: %v", err))
	}

	authDb, err := database.ConnectPostgres(os.Getenv("AUTH_SERVICE_DB"))
	if err != nil {
		panic("Failed to connect to Auth Database.")
	}

	authDb.AutoMigrate(authmodels.UserCredential{})
	authDb.AutoMigrate(authmodels.RefreshToken{})
	fmt.Println("Migrated Auth Database")

	userDb, err := database.ConnectPostgres(os.Getenv("USER_SERVICE_DB"))
	if err != nil {
		panic("Failed to connect to User Database.")
	}
	userDb.AutoMigrate(usersmodels.User{})
	userDb.AutoMigrate(usersmodels.Organization{})
	userDb.AutoMigrate(usersmodels.OrganizationMember{})
	userDb.AutoMigrate(usersmodels.OrganizationInvitation{})
	userDb.AutoMigrate(usersmodels.OrganizerApplication{})
	fmt.Println("Migrated User Database")

	eventDb, err := database.ConnectPostgres(os.Getenv("EVENT_SERVICE_DB"))
	if err != nil {
		panic("Failed to connect to Event Database.")
	}
	eventDb.AutoMigrate(eventmodels.Category{})
	fmt.Println("Migrated Event Database")

	if *seed {
		if err := seeders.SeedUsers(authDb, userDb, 20); err != nil {
			panic(fmt.Sprintf("Failed to seed users: %v", err))
		}
		fmt.Println("Seeded Users")

		if err := seeders.SeedCategories(eventDb); err != nil {
			panic(fmt.Sprintf("Failed to seed categories: %v", err))
		}
		fmt.Println("Seeded Categories")
	}
}
