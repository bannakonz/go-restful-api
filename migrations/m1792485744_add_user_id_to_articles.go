package migrations

import (
	"go-restful-api/models"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func m1792485744AddUserIDToArticle() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "1792485744",

		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Article{})
		},

		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropColumn(
				&models.Article{},
				"user_id",
			)
		},
	}
}
