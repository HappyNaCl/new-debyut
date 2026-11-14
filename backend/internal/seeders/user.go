package seeders

import (
	"fmt"
	"math/rand"
	"time"

	authmodels "github.com/HappyNaCl/new-debyut/internal/auth/models"
	"github.com/HappyNaCl/new-debyut/internal/common/snowflake"
	usersmodels "github.com/HappyNaCl/new-debyut/internal/users/models"
	faker "github.com/go-faker/faker/v4"
	"gorm.io/gorm"
)

type fakePerson struct {
	FirstName string `faker:"first_name"`
	LastName  string `faker:"last_name"`
	Email     string `faker:"email"`
}

func SeedUsers(authDb *gorm.DB, userDb *gorm.DB, count int) error {
	admin := usersmodels.User{
		FirstName:       "debyut",
		Surname:         "Admin",
		DateOfBirth:     time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		PhoneNumber:     "+6281200000000",
		Gender:          usersmodels.GenderMale,
		TermsAcceptedAt: time.Now().UTC(),
	}
	if _, err := seedUser(authDb, userDb, "admin@debyut.id", authmodels.RoleAdmin, admin); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}

	organizer := usersmodels.User{
		FirstName:       "debyut",
		Surname:         "Organizer",
		DateOfBirth:     time.Date(1992, 1, 1, 0, 0, 0, 0, time.UTC),
		PhoneNumber:     "+6281200000001",
		Gender:          usersmodels.GenderFemale,
		TermsAcceptedAt: time.Now().UTC(),
	}
	organizerId, err := seedUser(authDb, userDb, "organizer@debyut.id", authmodels.RoleOrganizer, organizer)
	if err != nil {
		return fmt.Errorf("seed organizer: %w", err)
	}
	if _, err := seedOrganization(userDb, organizerId, usersmodels.Organization{
		Name:        "debyut Live",
		Email:       "events@debyut.id",
		Description: "Sample organizer for local development.",
	}); err != nil {
		return fmt.Errorf("seed organization: %w", err)
	}

	genders := []usersmodels.Gender{usersmodels.GenderMale, usersmodels.GenderFemale}
	for index := range count {
		var person fakePerson
		if err := faker.FakeData(&person); err != nil {
			return fmt.Errorf("generate fake data at index %d: %w", index, err)
		}

		profile := usersmodels.User{
			FirstName:           person.FirstName,
			Surname:             person.LastName,
			DateOfBirth:         randomDateOfBirth(),
			PhoneNumber:         fmt.Sprintf("+628%010d", rand.Int63n(10_000_000_000)),
			Gender:              genders[rand.Intn(len(genders))],
			SubscribeNewsletter: rand.Intn(2) == 0,
			TermsAcceptedAt:     time.Now().UTC(),
		}
		if _, err := seedUser(authDb, userDb, person.Email, authmodels.RoleUser, profile); err != nil {
			return fmt.Errorf("seed user at index %d: %w", index, err)
		}
	}

	return nil
}

func seedUser(authDb *gorm.DB, userDb *gorm.DB, email string, role authmodels.Role, profile usersmodels.User) (string, error) {
	var credential authmodels.UserCredential
	if err := authDb.Where(authmodels.UserCredential{Email: email}).
		Attrs(authmodels.UserCredential{Id: snowflake.Generate(), Role: role}).
		FirstOrCreate(&credential).Error; err != nil {
		return "", err
	}

	return credential.Id, userDb.Where(usersmodels.User{Id: credential.Id}).
		Attrs(profile).
		FirstOrCreate(&usersmodels.User{}).Error
}

// seedOrganization creates an organization owned by ownerId, unless one with
// the same email exists, and returns its id.
func seedOrganization(userDb *gorm.DB, ownerId string, organization usersmodels.Organization) (string, error) {
	var existing usersmodels.Organization
	if err := userDb.Where("email = ?", organization.Email).First(&existing).Error; err == nil {
		return existing.Id, nil
	}

	organization.Id = snowflake.Generate()
	return organization.Id, userDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&organization).Error; err != nil {
			return err
		}
		return tx.Omit("Organization").Create(&usersmodels.OrganizationMember{
			OrganizationId: organization.Id,
			UserId:         ownerId,
			Role:           usersmodels.MemberRoleOwner,
			Status:         usersmodels.MemberStatusActive,
		}).Error
	})
}

func randomDateOfBirth() time.Time {
	age := 18 + rand.Intn(42)
	return time.Now().AddDate(-age, -rand.Intn(12), -rand.Intn(28)+1)
}
