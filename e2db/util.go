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

func (c *Connect) Count(ctx context.Context, model interface{}, useRO bool, query interface{}, args ...interface{}) (sql.NullInt64, error) {
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

func (c *Connect) DropTables(ctx context.Context, tables ...interface{}) error {
	for _, table := range tables {
		err := c.RW().WithContext(ctx).Migrator().DropTable(table)
		if err != nil {
			logrus.Errorf("failed to drop table %v: %v", table, err)
			return fmt.Errorf("failed to drop table %v: %w", table, err)
		}
	}
	return nil
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
	if modelValue.Kind() == reflect.Ptr {
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
