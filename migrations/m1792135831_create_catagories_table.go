package migrations

import (
    "go-restful-api/models"

    "github.com/go-gormigrate/gormigrate/v2"
    "gorm.io/gorm"

)

func m1792135831CreateCategoriesTable() *gormigrate.Migration {
    return &gormigrate.Migration{
        ID: "1792135831",

        Migrate: func(tx *gorm.DB) error {
            return tx.AutoMigrate(&models.Category{})
        },

        Rollback: func(tx *gorm.DB) error {
            return tx.Migrator().DropTable("categories")
        },
    }
}