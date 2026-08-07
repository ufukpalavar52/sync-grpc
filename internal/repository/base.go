package repository

import (
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"
	"gorm.io/gorm"
)

type BaseRepository struct {
	db *gorm.DB
}

func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db}
}

func (r *BaseRepository) ApplySmartFilters(filter interface{}, toSnake bool) *gorm.DB {
	db := r.db
	val := reflect.ValueOf(filter)
	typ := reflect.TypeOf(filter)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if field.IsZero() {
			continue
		}
		columnName := fieldType.Tag.Get("name")
		if columnName == "" || columnName == "-" {
			columnName = fieldType.Name
		}

		if toSnake {
			columnName = strcase.ToSnake(columnName)
		}

		filterType := fieldType.Tag.Get("filter")

		switch filterType {
		case "like":
			query := fmt.Sprintf("%s LIKE ?", fmt.Sprintf("%v", columnName))
			likeVal := fmt.Sprintf("%%%s%%", fmt.Sprintf("%v", field.Interface()))
			db = db.Where(query, likeVal)
		default:
			query := fmt.Sprintf("%s = ?", columnName)
			db = db.Where(query, field.Interface())
		}
	}

	return db
}
