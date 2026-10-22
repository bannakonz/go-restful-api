package migrations

import (
	"go-restful-api/models"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func m1792383343AddCategoryIdToArticle() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "1792383343",

		Migrate: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&models.Article{})

			var articles []models.Article
			tx.Unscoped().Find(&articles)
			for _, article := range articles {
				article.CategoryID = 2
				tx.Save(&article)
			}

			return err

		},

		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropColumn(
				&models.Article{},
				"category_id",
			)
		},
	}
}
