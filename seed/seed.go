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
		&models.User{},
		&models.Category{},
		&models.Article{},
		"migrations",
	); err != nil {
		log.Fatal(err)
	}

	// 2.
	migrations.Migrate()

	// Add Admin
	log.Println("Creating admin ...")

	admin := models.User{
		Email:    "admin@bannakon.com",
		Password: "passw0rd",
		Name:     "MR. Admin",
		Role:     "Admin",
		Avatar:   "https://picsum.photos/200",
	}
	hashedPassword, err := admin.GenerateEncryptedPassword()
	if err != nil {
		log.Fatal(err)
	}

	admin.Password = hashedPassword

	if err := db.Create(&admin).Error; err != nil {
		log.Fatal(err)
	}

	// Add normal users
	log.Println("Creating users...")
	numOfUsers := 50
	users := make([]models.User, 0, numOfUsers)
	userRoles := [2]string{"Editor", "Member"}

	for i := 1; i <= numOfUsers; i++ {
		user := models.User{
			Name:     faker.Name(),
			Email:    faker.Email(),
			Password: "passw0rd",
			Avatar:   "https://picsum.photos/seed/user-" + strconv.Itoa(i) + "/300/200",
			Role:     userRoles[rand.IntN(2)],
		}

		hashedPassword, err := user.GenerateEncryptedPassword()
		if err != nil {
			log.Fatal(err)
		}
		user.Password = hashedPassword

		if err := db.Create(&user).Error; err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}

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
			UserID:     users[rand.IntN(len(users))].ID,
		}

		if err := db.Create(&article).Error; err != nil {
			log.Fatal(err)
		}

		articles = append(articles, article)
	}

	log.Println("Seeding completed!")

}
