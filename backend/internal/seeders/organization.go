package seeders

import (
	"fmt"
	"time"

	authmodels "github.com/HappyNaCl/new-debyut/internal/auth/models"
	usersmodels "github.com/HappyNaCl/new-debyut/internal/users/models"
	"gorm.io/gorm"
)

type sampleOrganization struct {
	OwnerEmail string
	OwnerFirst string
	OwnerLast  string
	Phone      string
	Gender     usersmodels.Gender
	usersmodels.Organization
}

// sampleOrganizations each get an organizer account that owns them; sign in
// with the owner email to manage their events.
var sampleOrganizations = []sampleOrganization{
	{"organizer.nusantara@debyut.id", "Rina", "Wijaya", "+6281200000101", usersmodels.GenderFemale, usersmodels.Organization{
		Name: "Nusantara Sound", Email: "hello@nusantarasound.id",
		Description: "Concerts and music festivals across Indonesia.",
	}},
	{"organizer.ruang@debyut.id", "Bima", "Saputra", "+6281200000102", usersmodels.GenderMale, usersmodels.Organization{
		Name: "Ruang Kreatif", Email: "halo@ruangkreatif.id",
		Description: "Workshops, exhibitions and talks for creative people.",
	}},
	{"organizer.tawa@debyut.id", "Dewi", "Lestari", "+6281200000103", usersmodels.GenderFemale, usersmodels.Organization{
		Name: "Tawa Bersama", Email: "booking@tawabersama.id",
		Description: "Stand-up comedy nights and improv shows.",
	}},
	{"organizer.lari@debyut.id", "Andi", "Pratama", "+6281200000104", usersmodels.GenderMale, usersmodels.Organization{
		Name: "Lari Kota", Email: "info@larikota.id",
		Description: "Fun runs, marathons and community sports events.",
	}},
	{"organizer.tekno@debyut.id", "Sari", "Kusuma", "+6281200000105", usersmodels.GenderFemale, usersmodels.Organization{
		Name: "Tekno Meetup", Email: "team@teknomeetup.id",
		Description: "Tech conferences, hackathons and developer meetups.",
	}},
}

type SeededOrganization struct {
	Id      string
	OwnerId string
}

// SeedOrganizations creates the sample organizations and their owners and
// returns every seeded organization, including "debyut Live" from SeedUsers.
// Existing organizations (matched by email) are reused.
func SeedOrganizations(authDb *gorm.DB, userDb *gorm.DB) ([]SeededOrganization, error) {
	var seeded []SeededOrganization

	var debyutLive usersmodels.OrganizationMember
	err := userDb.Joins("JOIN organizations ON organizations.id = organization_members.organization_id").
		Where("organizations.email = ? AND organization_members.role = ?", "events@debyut.id", usersmodels.MemberRoleOwner).
		First(&debyutLive).Error
	if err == nil {
		seeded = append(seeded, SeededOrganization{Id: debyutLive.OrganizationId, OwnerId: debyutLive.UserId})
	}

	for _, sample := range sampleOrganizations {
		ownerId, err := seedUser(authDb, userDb, sample.OwnerEmail, authmodels.RoleOrganizer, usersmodels.User{
			FirstName:       sample.OwnerFirst,
			Surname:         sample.OwnerLast,
			DateOfBirth:     time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC),
			PhoneNumber:     sample.Phone,
			Gender:          sample.Gender,
			TermsAcceptedAt: time.Now().UTC(),
		})
		if err != nil {
			return nil, fmt.Errorf("seed owner of %s: %w", sample.Name, err)
		}
		id, err := seedOrganization(userDb, ownerId, sample.Organization)
		if err != nil {
			return nil, fmt.Errorf("seed organization %s: %w", sample.Name, err)
		}
		seeded = append(seeded, SeededOrganization{Id: id, OwnerId: ownerId})
	}
	return seeded, nil
}
