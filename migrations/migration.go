package migrations

import (
	"go-restful-api/config"
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
)

func Migrate() {
	db := config.GetDB()
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		m1791951215CreateArticlesTable(),
		m1792135831CreateCategoriesTable(),
		m1792383343AddCategoryIdToArticle(),
	})

	if err := m.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
