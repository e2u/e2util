package e2rest

import (
	"gorm.io/gorm"
)

func GetList(db *gorm.DB) {

}

func GetOne(db *gorm.DB) {

}

func GetMany(db *gorm.DB) {

}

func GetManyReference(db *gorm.DB) {

}

func Create[T any](db *gorm.DB, obj T) (T, error) {
	err := db.Create(obj).Error
	return obj, err
}

func Update[T any](db *gorm.DB, obj T) (T, error) {
	err := db.Updates(obj).Error
	return obj, err
}

func UpdateMany(db *gorm.DB) {

}
func Delete[T any](db *gorm.DB, obj T) error {
	return nil
}

func DeleteMany[T any](db *gorm.DB, objs []T) error {
	return nil
}
