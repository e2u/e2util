package e2db

import (
	"context"
	"database/sql"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	var r T
	dd.Scan(&r)
	return r, nil
}
