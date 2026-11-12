package seeders

import (
	"github.com/HappyNaCl/new-debyut/internal/common/snowflake"
	eventmodels "github.com/HappyNaCl/new-debyut/internal/event/models"
	"gorm.io/gorm"
)

var defaultCategories = []eventmodels.Category{
	{Slug: "art-culture", Name: "Art & Culture", Icon: "palette"},
	{Slug: "automotive", Name: "Automotive", Icon: "directions-car"},
	{Slug: "business-finance", Name: "Business & Finance", Icon: "business-center"},
	{Slug: "comedy-performance", Name: "Comedy & Performance", Icon: "theater-comedy"},
	{Slug: "education", Name: "Education", Icon: "school"},
	{Slug: "environment-sustainability", Name: "Environment & Sustainability", Icon: "park"},
	{Slug: "family-kids", Name: "Family & Kids", Icon: "family-restroom"},
	{Slug: "fashion-beauty", Name: "Fashion & Beauty", Icon: "checkroom"},
	{Slug: "food-drink", Name: "Food & Drink", Icon: "restaurant"},
	{Slug: "health-wellness", Name: "Health & Wellness", Icon: "health-and-safety"},
	{Slug: "hobbies-lifestyle", Name: "Hobbies & Lifestyle", Icon: "sports-esports"},
	{Slug: "media-entertainment", Name: "Media & Entertainment", Icon: "movie"},
	{Slug: "music", Name: "Music", Icon: "music-note"},
	{Slug: "personal-development", Name: "Personal Development", Icon: "self-improvement"},
	{Slug: "religion-spirituality", Name: "Religion & Spirituality", Icon: "volunteer-activism"},
	{Slug: "school-campus", Name: "School & Campus Activities", Icon: "backpack"},
	{Slug: "science-technology", Name: "Science & Technology", Icon: "science"},
	{Slug: "social-politics", Name: "Social & Politics", Icon: "how-to-vote"},
	{Slug: "sports-fitness", Name: "Sports & Fitness", Icon: "sports-soccer"},
	{Slug: "travel-outdoor", Name: "Travel & Outdoor", Icon: "luggage"},
}

func SeedCategories(eventDb *gorm.DB) error {
	for _, category := range defaultCategories {
		if err := eventDb.Where(eventmodels.Category{Slug: category.Slug}).
			Attrs(eventmodels.Category{Id: snowflake.Generate(), Name: category.Name, Icon: category.Icon}).
			FirstOrCreate(&eventmodels.Category{}).Error; err != nil {
			return err
		}
	}
	return nil
}
