package e2db

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (c *Connect) Count(ctx context.Context, model any, useRO bool, query any, args ...any) (sql.NullInt64, error) {
	db := c.RW()
	if useRO {
		db = c.RO()
	}
	var count int64
	err := db.WithContext(ctx).Model(model).Where(query, args...).Count(&count).Error
	if err != nil {
		logrus.Errorf("count error=%v", err)
		return sql.NullInt64{Int64: 0, Valid: false}, err
	}
	return sql.NullInt64{Int64: count, Valid: true}, nil
}

func (c *Connect) DropTables(ctx context.Context, tables ...any) error {
	for _, table := range tables {
		err := c.RW().WithContext(ctx).Migrator().DropTable(table)
		if err != nil {
			logrus.Errorf("failed to drop table %v: %v", table, err)
			return fmt.Errorf("failed to drop table %v: %w", table, err)
		}
	}
	return nil
}

func (c *Connect) DropCascadTables(ctx context.Context, tables ...any) error {
	return c.RW().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, table := range tables {
			err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %v CASCADE", table)).Error
			if err != nil {
				tx.Rollback()
				logrus.Errorf("failed to drop table %v: %v", table, err)
				return fmt.Errorf("failed to drop table %v: %w", table, err)
			}
		}
		return nil
	})
}

func (c *Connect) Truncate(ctx context.Context, cascad bool, tables ...any) error {
	cascadStr := ""
	if cascad {
		cascadStr = "CASCADE"
	}
	return c.RW().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, table := range tables {
			if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %v %s", table, cascadStr)).Error; err != nil {
				tx.Rollback()
				logrus.Errorf("failed to truncate table %v: %v", table, err)
				return fmt.Errorf("failed to truncate table %v: %w", table, err)
			}
		}
		return nil
	})
}

// DBHandler is a generic helper struct for database operations.
type DBHandler[T any] struct {
	RW *gorm.DB
	RO *gorm.DB
}

// SaveAndPreload saves and preloads the model.
func (h *DBHandler[T]) SaveAndPreload(ctx context.Context, model T) (T, error) {
	// Save the model using the RW connection
	tx := h.RW.WithContext(ctx).Save(model)
	if tx.Error != nil {
		return model, fmt.Errorf("failed to save model: %w", tx.Error)
	}

	// Get the schema to identify primary keys (e.g., PKAID)
	schema := tx.Statement.Schema
	if schema == nil || len(schema.PrimaryFields) == 0 {
		return model, fmt.Errorf("no schema or primary key defined for model")
	}

	// Build conditions based on primary key(s)
	conditions := make(map[string]any)
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Pointer {
		modelValue = modelValue.Elem() // Dereference if it's a pointer
	}
	for _, field := range schema.PrimaryFields {
		value, zero := field.ValueOf(ctx, modelValue)
		if zero {
			return model, fmt.Errorf("primary key %s is zero or unset", field.DBName)
		}
		conditions[field.DBName] = value
	}

	// Fetch the model with associations using the RO connection
	var result T
	err := h.RO.WithContext(ctx).Preload(clause.Associations).Where(conditions).First(&result).Error
	if err != nil {
		logrus.Warnf("failed to preload model after save: %v", err)
		return model, fmt.Errorf("failed to preload associations: %w", err)
	}

	return result, nil
}

// NewDBHandler creates a new DBHandler for the given type and Connect instance.
func NewDBHandler[T any](c *Connect) *DBHandler[T] {
	return &DBHandler[T]{RW: c.RW(), RO: c.RO()}
}

func FixPGSequenceValue(db *gorm.DB, tableName string, column string) error {
	var sequenceName string
	err := db.Raw(
		"SELECT pg_get_serial_sequence($1, $2)",
		tableName,
		column,
	).Scan(&sequenceName).Error
	if err != nil {
		return fmt.Errorf("failed to get sequence name: %w", err)
	}

	if sequenceName == "" {
		return fmt.Errorf("no sequence found for table %s column %s", tableName, column)
	}

	query := fmt.Sprintf(
		"SELECT setval('%s', COALESCE((SELECT MAX(%s) FROM %s), 0) + 1, false)",
		sequenceName,
		column,
		tableName,
	)

	if err := db.Exec(query).Error; err != nil {
		return fmt.Errorf("failed to update sequence value: %w", err)
	}

	return nil
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func MakeUpdates(um map[string]any, exists, update any, mapKey string, allowNil bool) {
	if update == nil {
		if allowNil && !isNil(exists) {
			um[mapKey] = nil
		}
		return
	}
	if !reflect.DeepEqual(update, exists) {
		um[mapKey] = update
	}
}

func MakeTimeUpdate(um map[string]any, exists, update time.Time, mapKey string) {
	if exists.IsZero() && update.IsZero() {
		return
	}
	if !exists.Equal(update) {
		um[mapKey] = update
	}
}

func MakeFloatUpdate(um map[string]any, exists, update float64, mapKey string, epsilon float64) {
	if math.IsNaN(exists) != math.IsNaN(update) {
		um[mapKey] = update
		return
	}
	if math.IsNaN(exists) && math.IsNaN(update) {
		return
	}
	if math.Abs(exists-update) > epsilon {
		um[mapKey] = update
	}
}

func MakePtrFloatUpdate(um map[string]any, exists, update *float64, mapKey string, epsilon float64, allowNil bool) {
	if exists == nil || update == nil {
		MakeUpdates(um, exists, update, mapKey, allowNil)
		return
	}
	if math.Abs(*exists-*update) > epsilon {
		um[mapKey] = update
	}
}

func MakeDecimalUpdate(um map[string]any, exists, update decimal.Decimal, mapKey string) {
	// 使用 Equal（底層做數值比較，1.0 與 1.00 會被視為相等）
	if !exists.Equal(update) {
		um[mapKey] = update
	}
}

// 指針 decimal（可為 nil），allowNil 控制是否允許把值設為 NULL
func MakePtrDecimalUpdate(um map[string]any, exists, update *decimal.Decimal, mapKey string, allowNil bool) {
	switch {
	case exists == nil && update == nil:
		return
	case exists == nil && update != nil:
		um[mapKey] = update
		return
	case exists != nil && update == nil:
		if allowNil {
			um[mapKey] = nil
		}
		return
	default:
		// 兩者皆非 nil，用 Equal 比較數值
		if !exists.Equal(*update) {
			um[mapKey] = update
		}
	}
}
