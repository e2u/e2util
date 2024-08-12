package e2db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/e2u/e2util/e2map"
	"gorm.io/gorm"
)

type ModelWithSoftDelete struct {
	PKAID     uint           `gorm:"primarykey;column:pkaid" json:"pkaid"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Model struct {
	PKAID     uint      `gorm:"primarykey;column:pkaid" json:"pkaid"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
}

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
//	Quantity      int             `gorm:"column:quantity" json:"quantity"`               // Amount of  product,like 1 box 24 services etc
//	UnitOfMeasure string          `gorm:"column:unit_of_measure" json:"unit_of_measure"` // The unit of measurement for the product (e.g., each, box, kilogram, liter).
//	Description   string          `gorm:"column:description" json:"description"`
//	Pictures      JSONBArray `gorm:"column:pictures;type:jsonb" json:"pictures"`
//	Attributes    JSONBMap   `gorm:"column:attributes;type:jsonb" json:"attributes"` // Use JSON storage the attributes
//  Extra         *Extra     `gorm:"column:extra;type:jsonb" json:"extra"`
//}

type JSONBArray []any

func (jsonField JSONBArray) Value() (driver.Value, error) {
	return json.Marshal(jsonField)
}

func (jsonField *JSONBArray) Scan(value any) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(data, &jsonField)
}

type JSONBMap map[string]any

func (jsonField JSONBMap) Value() (driver.Value, error) {
	return json.Marshal(jsonField)
}

func (jsonField *JSONBMap) Scan(value any) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(data, &jsonField)
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
