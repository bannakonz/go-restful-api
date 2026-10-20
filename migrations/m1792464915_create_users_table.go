package migrations

import (
	"go-restful-api/models"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func m1792464915CreateUsersTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "1792464915",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.User{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("users")
		},
	}

}
