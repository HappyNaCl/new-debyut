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
	"gorm.io/gorm"
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

	normalizeEmails(authDb, "user_credentials", "email")
	authDb.AutoMigrate(authmodels.UserCredential{})
	authDb.AutoMigrate(authmodels.RefreshToken{})
	fmt.Println("Migrated Auth Database")

	userDb, err := database.ConnectPostgres(os.Getenv("USER_SERVICE_DB"))
	if err != nil {
		panic("Failed to connect to User Database.")
	}
	normalizeEmails(userDb, "organizations", "email")
	normalizeEmails(userDb, "organizer_applications", "organization_email")
	normalizeEmails(userDb, "organization_invitations", "email")
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

// normalizeEmails rewrites existing emails to the stored form (trimmed and
// lowercased, see emails.Normalize) so plain equality lookups and unique
// indexes work. It fails loudly if two rows differ only by case, since those
// must be merged by hand.
func normalizeEmails(db *gorm.DB, table, column string) {
	if !db.Migrator().HasTable(table) {
		return
	}
	result := db.Exec(fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = lower(trim(%[2]s)) WHERE %[2]s <> lower(trim(%[2]s))", table, column,
	))
	if result.Error != nil {
		panic(fmt.Sprintf("Failed to normalize %s.%s (rows that differ only by case must be merged first): %v", table, column, result.Error))
	}
	if result.RowsAffected > 0 {
		fmt.Printf("Normalized %d email(s) in %s.%s\n", result.RowsAffected, table, column)
	}
}
