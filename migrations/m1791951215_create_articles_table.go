package migrations

import (
	"go-restful-api/models"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func m1791951215CreateArticlesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "1791951215",

		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Article{})
		},

		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("articles")
		},
	}
}
