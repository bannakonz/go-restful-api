package seed

import (
	"go-restful-api/config"
	"go-restful-api/migrations"
	"go-restful-api/models"
	"log"
	"math/rand/v2"
	"strconv"

	"github.com/go-faker/faker/v4"
)

func Load() {
	db := config.GetDB()

	// 1. Drop tables
	if err := db.Migrator().DropTable(
		&models.Category{},
		&models.Article{},
		"migrations",
	); err != nil {
		log.Fatal(err)
	}

	// 2.
	migrations.Migrate()

	// 3. Create categories
	log.Println("Creating categories...")
	numOfCategories := 20

	categories := make([]models.Category, 0, numOfCategories)

	for i := 1; i <= numOfCategories; i++ {
		category := models.Category{
			Name:        faker.Word(),
			Description: faker.Paragraph(),
		}

		if err := db.Create(&category).Error; err != nil {
			log.Fatal(err)
		}

		categories = append(categories, category)
	}

	// 4. Create articles
	log.Println("Creating articles...")

	numOfArticles := 50
	articles := make([]models.Article, 0, numOfArticles)

	for i := 1; i <= numOfArticles; i++ {
		article := models.Article{
			Title:      faker.Sentence() + " #" + strconv.Itoa(i),
			Body:       faker.Paragraph(),
			Image:      "https://picsum.photos/seed/article-" + strconv.Itoa(i) + "/300/200",
			CategoryId: categories[rand.IntN(len(categories))].ID,
		}

		if err := db.Create(&article).Error; err != nil {
			log.Fatal(err)
		}

		articles = append(articles, article)
	}

	log.Println("Seeding completed!")

}
