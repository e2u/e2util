package e2db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func MustGetCount(db *gorm.DB, model interface{}, query interface{}, args ...interface{}) sql.NullInt64 {
	var c int64
	if err := db.Model(model).Where(query, args).Count(&c).Error; err != nil {
		logrus.Errorf("MustGetCount error=%v", err)
		return sql.NullInt64{
			Int64: 0,
			Valid: false,
		}
	}
	return sql.NullInt64{
		Int64: c,
		Valid: true,
	}
}

func SaveWithContext[T any](ctx context.Context, db *gorm.DB, model T) (T, error) {
	dd := db.WithContext(ctx).Save(model)
	if err := dd.Error; err != nil {
		return model, err
	}
	qm := make(map[string]any)
	for _, pn := range dd.Statement.Schema.PrimaryFields {
		if !pn.PrimaryKey {
			continue
		}
		r := reflect.ValueOf(model)
		f := reflect.Indirect(r).FieldByName(pn.Name)
		switch {
		case f.CanFloat():
			qm[pn.DBName] = f.Float()
		case f.CanInt():
			qm[pn.DBName] = f.Int()
		case f.CanUint():
			qm[pn.DBName] = f.Uint()
		default:
			qm[pn.DBName] = fmt.Sprintf("%v", pn.DBName)
		}
	}
	var rv T
	err := dd.Preload(clause.Associations).Limit(1).Where(qm).Find(&rv).Error
	return rv, err
}
