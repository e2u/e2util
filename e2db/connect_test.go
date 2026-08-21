package e2db

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/e2u/e2util/e2logrus"
	"github.com/e2u/e2util/e2model"
	"github.com/sirupsen/logrus"
)

// 測試用的模型
type Product struct {
	PKAID      uint      `gorm:"primarykey;column:pkaid" json:"pkaid"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
	UpdatedAt  time.Time `gorm:"index" json:"updated_at"`
	Name       string    `gorm:"column:name" json:"name"`
	Category   Category  `gorm:"foreignKey:CategoryID"`
	CategoryID uint      `gorm:"column:category_id"`
}

type Category struct {
	PKAID     uint      `gorm:"primarykey;column:pkaid" json:"pkaid"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
	Name      string    `gorm:"column:name" json:"name"`
}

// 測試初始化函數
func setupTestDB(t *testing.T) *Connect {
	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	cfg := &Config{
		Driver: "sqlite",
		Writer: dsn,
		Readers: []string{
			dsn,
		},
		EnableMigrate:       true,
		SQLLogSlowThreshold: 100,
		SQLLogColorful:      false,
		LoggerConfig: &e2logrus.Config{
			Level:  "info",
			Format: "text",
		},
	}
	conn, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create test DB connection: %v", err)
	}
	return conn
}

// TestNew 測試 New 函數
func TestNew(t *testing.T) {
	conn := setupTestDB(t)
	if conn.db == nil {
		t.Error("expected db to be initialized, got nil")
	}
	if len(conn.roDb) != 1 {
		t.Errorf("expected 1 read-only connection, got %d", len(conn.roDb))
	}
}

// TestRWRO 測試 RW 和 RO 方法
func TestRWRO(t *testing.T) {
	conn := setupTestDB(t)

	rwDB := conn.RW()
	if rwDB == nil {
		t.Error("RW() returned nil")
	}

	roDB := conn.RO()
	if roDB == nil {
		t.Error("RO() returned nil")
	}

	rwDebug := conn.RW(&Option{Debug: true})
	if rwDebug == nil {
		t.Error("RW() with debug returned nil")
	}
}

// TestExists 測試 Exists 方法
func TestExists(t *testing.T) {
	conn := setupTestDB(t)
	if err := conn.AutoMigrate(context.Background(), &Product{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	product := &Product{Name: "Test Product"}
	if err := conn.RW().Create(product).Error; err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	exists := conn.Exists(&Product{}, "name = ?", true, "Test Product")
	if exists.Error != nil {
		t.Errorf("Exists failed: %v", exists.Error)
	}
	if !exists.Bool {
		t.Error("expected product to exist, got false")
	}
	if !exists.Valid {
		t.Error("expected valid to be true when no error")
	}

	exists = conn.Exists(&Product{}, "name = ?", true, "Nonexistent")
	if exists.Error != nil {
		t.Errorf("Exists failed: %v", exists.Error)
	}
	if exists.Bool {
		t.Error("expected product not to exist, got true")
	}
	if !exists.Valid {
		t.Error("expected valid to be true when no error")
	}
}

// TestPatch 測試 Patch 方法
func TestPatch(t *testing.T) {
	conn := setupTestDB(t)
	if err := conn.AutoMigrate(context.Background(), &Product{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	product := &Product{Name: "Old Name"}
	if err := conn.RW().Create(product).Error; err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	patchs := []*e2model.HttpPatch{
		{
			Op:    e2model.HttpPatchOpReplace,
			Path:  "name",
			Value: "New Name",
		},
	}
	err := conn.Patch(context.Background(), product, patchs)
	if err != nil {
		t.Errorf("Patch failed: %v", err)
	}

	var updated Product
	if err := conn.RW().First(&updated, product.PKAID).Error; err != nil {
		t.Errorf("failed to fetch updated product: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected name to be 'New Name', got '%s'", updated.Name)
	}

	invalidPatch := []*e2model.HttpPatch{
		{
			Op:    "invalid_op",
			Path:  "name",
			Value: "Invalid",
		},
	}
	err = conn.Patch(context.Background(), product, invalidPatch)
	if err != nil {
		t.Errorf("Patch with invalid op failed unexpectedly: %v", err)
	}

	emptyPathPatch := []*e2model.HttpPatch{
		{
			Op:    e2model.HttpPatchOpReplace,
			Path:  "",
			Value: "Should Fail",
		},
	}
	err = conn.Patch(context.Background(), product, emptyPathPatch)
	if err == nil || err.Error() != "invalid patch: empty path" {
		t.Errorf("expected error for empty path, got: %v", err)
	}
}

// TestAutoMigrate 測試 AutoMigrate 方法
func TestAutoMigrate(t *testing.T) {
	conn := setupTestDB(t)
	err := conn.AutoMigrate(context.Background(), &Product{}, &Category{})
	if err != nil {
		t.Errorf("AutoMigrate failed: %v", err)
	}

	var count int64
	if err := conn.RW().Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='products'").Scan(&count).Error; err != nil {
		t.Errorf("failed to check table existence: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 'products' table to exist, got count %d", count)
	}
}

func TestAutoMigrateDisabled(t *testing.T) {
	conn := setupTestDB(t)
	conn.EnableMigrate = false
	if err := conn.AutoMigrate(context.Background(), &Product{}); err != nil {
		t.Fatalf("AutoMigrate with migrate disabled should no-op, got: %v", err)
	}

	var count int64
	if err := conn.RW().Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='products'").Scan(&count).Error; err != nil {
		t.Fatalf("failed to check table existence: %v", err)
	}
	if count != 0 {
		t.Errorf("expected no 'products' table when migrate is disabled, got count %d", count)
	}
}

// TestCreateSchema 測試 CreateSchema 方法
func TestCreateSchema(t *testing.T) {
	conn := setupTestDB(t)
	err := conn.CreateSchema(context.Background(), "test_schema")
	if err == nil || err.Error() != "CreateSchema is only supported for PostgreSQL" {
		t.Errorf("expected error for SQLite, got: %v", err)
	}
}

// TestCount 測試 Count 方法
func TestCount(t *testing.T) {
	conn := setupTestDB(t)
	if err := conn.AutoMigrate(context.Background(), &Product{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	products := []Product{{Name: "P1"}, {Name: "P2"}}
	if err := conn.RW().Create(&products).Error; err != nil {
		t.Fatalf("failed to create products: %v", err)
	}

	count, err := conn.Count(context.Background(), &Product{}, true, "name LIKE ?", "P%")
	if err != nil {
		t.Errorf("Count failed: %v", err)
	}
	if !count.Valid || count.Int64 != 2 {
		t.Errorf("expected count 2, got %v", count)
	}
}

// TestSaveAndPreload 測試 SaveAndPreload 方法
func TestSaveAndPreload(t *testing.T) {
	conn := setupTestDB(t)
	if err := conn.AutoMigrate(context.Background(), &Product{}, &Category{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// 使用指針類型
	product := &Product{
		Name:     "Test Product",
		Category: Category{Name: "Electronics"},
	}

	handler := NewDBHandler[*Product](conn)
	result, err := handler.SaveAndPreload(context.Background(), product)
	if err != nil {
		t.Errorf("SaveAndPreload failed: %v", err)
	}

	if result.PKAID == 0 {
		t.Error("expected PKAID to be set, got 0")
	}
	if result.Category.Name != "Electronics" {
		t.Errorf("expected category name 'Electronics', got '%s'", result.Category.Name)
	}
}

// TestDropTables 測試 DropTables 方法
func TestDropTables(t *testing.T) {
	conn := setupTestDB(t)
	if err := conn.AutoMigrate(context.Background(), &Product{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	if err := conn.RW().Create(&Product{Name: "Test"}).Error; err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	err := conn.DropTables(context.Background(), &Product{})
	if err != nil {
		t.Errorf("DropTables failed: %v", err)
	}

	var count int64
	if err := conn.RW().Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='products'").Scan(&count).Error; err != nil {
		t.Errorf("failed to check table existence: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 'products' table to be dropped, got count %d", count)
	}
}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.InfoLevel)
	m.Run()
}
