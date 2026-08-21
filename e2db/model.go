package e2db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/e2u/e2util/e2map"
	"gorm.io/gorm"
)

// Model represents a basic database model with a physical primary key and timestamps.
// Model 表示具有物理主鍵和時間戳的基本資料庫模型。
type Model struct {
	PKAID uint `gorm:"primarykey;column:pkaid;autoIncrement" json:"pkaid"` // Physical auto-incremented primary key
	// 物理自動遞增主鍵
	CreatedAt time.Time `gorm:"index" json:"created_at"` // Record creation timestamp
	// 記錄創建時間戳
	UpdatedAt time.Time `gorm:"index" json:"updated_at"` // Record last update timestamp
	// 記錄最後更新時間戳
}

// ModelWithSoftDelete extends Model with soft delete functionality.
// ModelWithSoftDelete 擴展 Model 並添加軟刪除功能。
type ModelWithSoftDelete struct {
	Model
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"` // Soft delete timestamp
	// 軟刪除時間戳
}

// type ModelWithSoftDelete struct {
//	PKAID     uint           `gorm:"primarykey;column:pkaid" json:"pkaid"`
//	CreatedAt time.Time      `gorm:"index" json:"created_at"`
//	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
//	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
//}
//
// type Model struct {
//	PKAID     uint      `gorm:"primarykey;column:pkaid" json:"pkaid"`
//	CreatedAt time.Time `gorm:"index" json:"created_at"`
//	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
//}

// Example
// type Extra struct {
//	F1 string
//	F2 string
//}
// func (t Extra) Value() (driver.Value, error) {
//	return JSONBValue(t)
//}
//
// func (t *Extra) Scan(value any) error {
//	return JSONBScan(t, value)
//}

// type Product struct {
//	*Model
//	Key            string          `gorm:"column:id" json:"id"`
//	Name          string          `gorm:"column:name" json:"name"`
//	Brand         string          `gorm:"column:brand" json:"brand"`
//	UPC           string          `gorm:"column:upc" json:"upc"`                         // Universal Product Code (UPC)
//	// 通用產品代碼 (UPC)
//	Quantity      int             `gorm:"column:quantity" json:"quantity"`               // Amount of  product,like 1 box 24 services etc
// 產品數量，例如 1 盒 24 份等
//	UnitOfMeasure string          `gorm:"column:unit_of_measure" json:"unit_of_measure"` // The unit of measurement for the product (e.g., each, box, kilogram, liter).
// 產品的計量單位（例如每個、盒、公斤、升）。
//	Description   string          `gorm:"column:description" json:"description"`
//	Pictures      JSONBArray `gorm:"column:pictures;type:jsonb" json:"pictures"`
//	Attributes    JSONBMap   `gorm:"column:attributes;type:jsonb" json:"attributes"` // Use JSON storage the attributes
// 使用 JSON 儲存屬性
//  Extra         *Extra     `gorm:"column:extra;type:jsonb" json:"extra"`
//}

type JSONBHandler[T any] struct {
	Data T `json:"data"`
}

func (j JSONBHandler[T]) Value() (driver.Value, error) {
	return json.Marshal(j.Data)
}

func (j *JSONBHandler[T]) Scan(value any) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(data, &j.Data)
}

// JSONBArray is a slice of arbitrary values stored as JSONB.
// JSONBArray 是儲存為 JSONB 的任意值切片。
type JSONBArray JSONBHandler[[]any]

// type JSONBArray []any
func (jsonField JSONBArray) Value() (driver.Value, error) {
	if jsonField.Data == nil {
		return nil, nil // SQL NULL when Data is nil.
	}
	return json.Marshal(jsonField.Data)
}

// Scan deserializes a JSON array (e.g., ["a","b"]) or legacy format (e.g., {"data": ["a","b"]}) into the JSONBArray's Data field.
// Scan 將 JSON 陣列（例如 ["a","b"]）或舊格式（例如 {"data": ["a","b"]}）反序列化至 JSONBArray 的 Data 字段。
func (jsonField *JSONBArray) Scan(value any) error {
	if value == nil {
		jsonField.Data = nil // Set Data to nil for DB NULL.
		// 對於 DB NULL 將 Data 設為 nil。
		return nil
	}
	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	if string(data) == "null" {
		jsonField.Data = nil // Handle JSON "null" by setting Data to nil.
		// 處理 JSON "null" 將 Data 設為 nil。
		return nil
	}
	// First try unmarshaling as legacy wrapper for array
	// 先試圖 unmarshal 為陣列的舊 wrapper
	var wrapper struct {
		Data []any `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.Data != nil {
		jsonField.Data = wrapper.Data
		return nil
	}
	// Fallback to unmarshaling directly as []any
	// 回退直接 unmarshal 為 []any
	if err := json.Unmarshal(data, &jsonField.Data); err != nil {
		return fmt.Errorf("unmarshal JSONBArray: %w", err)
	}
	return nil
}

// MarshalJSON serializes the JSONBArray's Data field directly for JSON responses, e.g., ["a","b"].
// MarshalJSON 將 JSONBArray 的 Data 字段直接序列化為 JSON 回應，例如 ["a","b"]。
func (jsonField JSONBArray) MarshalJSON() ([]byte, error) {
	if jsonField.Data == nil {
		return json.Marshal([]any{})
	}
	return json.Marshal(jsonField.Data)
}

// UnmarshalJSON deserializes JSON data into the JSONBArray's Data field.
// UnmarshalJSON 將 JSON 資料反序列化至 JSONBArray 的 Data 字段。
func (jsonField *JSONBArray) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &jsonField.Data)
}

// JSONBMap is a map of string to arbitrary values stored as JSONB.
// JSONBMap 是儲存為 JSONB 的字串至任意值的映射。
type JSONBMap JSONBHandler[map[string]any]

// MarshalJSON serializes the JSONBMap's Data field directly for JSON responses, e.g., {"key": "value"}.
// MarshalJSON 將 JSONBMap 的 Data 字段直接序列化為 JSON 回應，例如 {"key": "value"}。
func (jsonField JSONBMap) MarshalJSON() ([]byte, error) {
	if jsonField.Data == nil {
		return json.Marshal(map[string]any{})
	}
	return json.Marshal(jsonField.Data)
}

// UnmarshalJSON deserializes JSON data into the JSONBMap's Data field.
// UnmarshalJSON 將 JSON 資料反序列化至 JSONBMap 的 Data 字段。
func (jsonField *JSONBMap) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &jsonField.Data)
}

// type JSONBMap map[string]any
func (jsonField JSONBMap) Value() (driver.Value, error) {
	if jsonField.Data == nil {
		return nil, nil // SQL NULL when Data is nil.
	}
	return json.Marshal(jsonField.Data)
}

// Scan deserializes a JSON object (e.g., {"key": "value"}) or legacy format (e.g., {"data": {"key": "value"}}) into the JSONBMap's Data field.
// Scan 將 JSON 物件（例如 {"key": "value"}）或舊格式（例如 {"data": {"key": "value"}}）反序列化至 JSONBMap 的 Data 字段。
func (jsonField *JSONBMap) Scan(value any) error {
	if value == nil {
		jsonField.Data = nil // Set Data to nil for DB NULL.
		// 對於 DB NULL 將 Data 設為 nil。
		return nil
	}
	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	if string(data) == "null" {
		jsonField.Data = nil // Handle JSON "null" by setting Data to nil.
		// 處理 JSON "null" 將 Data 設為 nil。
		return nil
	}
	// First try unmarshaling as legacy wrapper for map
	// 先試圖 unmarshal 為映射的舊 wrapper
	var wrapper struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.Data != nil {
		jsonField.Data = wrapper.Data
		return nil
	}
	// Fallback to unmarshaling directly as map[string]any
	// 回退直接 unmarshal 為 map[string]any
	if err := json.Unmarshal(data, &jsonField.Data); err != nil {
		return fmt.Errorf("unmarshal JSONBMap: %w", err)
	}
	return nil
}

type JSONB interface {
	Value() (driver.Value, error)
	Scan(value any) error
}

func JSONBValue[T any](model T) (driver.Value, error) {
	return json.Marshal(model)
}

func JSONBScan[T any](model T, value any) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(data, &model)
}

type JSONBMapArray []e2map.Map

func (jsonField JSONBMapArray) Value() (driver.Value, error) {
	return json.Marshal(jsonField)
}

func (jsonField *JSONBMapArray) Scan(value any) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(data, &jsonField)
}
