package e2db

import (
	"encoding/json"
	"testing"

	"github.com/e2u/e2util/e2map"
)

// TestJSONBArray tests the JSONBArray type's methods.
// TestJSONBArray 測試 JSONBArray 類型的方法。
func TestJSONBArray(t *testing.T) {
	// Test case: Empty array.
	// 測試案例：空陣列。
	t.Run("Empty", func(t *testing.T) {
		var arr JSONBArray
		val, err := arr.Value()
		if err != nil {
			t.Fatalf("Value() error: %v", err)
		}
		if val != nil {
			t.Errorf("Expected nil, got %v", val)
		}

		// MarshalJSON should return [].
		// MarshalJSON 應返回 []。
		marshaled, err := json.Marshal(arr)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}
		if string(marshaled) != "[]" {
			t.Errorf("Expected [], got %s", marshaled)
		}
	})

	// Test case: Nil data.
	// 測試案例：Nil 資料。
	t.Run("Nil", func(t *testing.T) {
		var arr JSONBArray // Use value type to get non-nil pointer. / 使用值類型以取得非 nil 指標。
		err := (&arr).Scan(nil)
		if err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if arr.Data != nil {
			t.Errorf("Expected Data nil, got %v", arr.Data)
		}
	})

	// Test case: Direct format array.
	// 測試案例：直接格式陣列。
	t.Run("DirectFormat", func(t *testing.T) {
		data := []byte(`["a", "b"]`)
		var arr JSONBArray
		err := arr.Scan(data)
		if err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if len(arr.Data) != 2 || arr.Data[0].(string) != "a" {
			t.Errorf("Expected [a b], got %v", arr.Data)
		}

		// Value should return the direct array.
		// Value 應返回直接陣列。
		val, _ := arr.Value()
		if string(val.([]byte)) != `["a","b"]` {
			t.Errorf("Expected [\"a\",\"b\"], got %s", val)
		}

		// MarshalJSON should return direct array.
		// MarshalJSON 應返回直接陣列。
		marshaled, _ := json.Marshal(arr)
		if string(marshaled) != `["a","b"]` {
			t.Errorf("Expected [\"a\",\"b\"], got %s", marshaled)
		}

		// UnmarshalJSON.
		// UnmarshalJSON。
		err = json.Unmarshal(data, &arr)
		if err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}
		if len(arr.Data) != 2 {
			t.Errorf("Expected length 2, got %d", len(arr.Data))
		}
	})

	// Test case: Legacy wrapped format.
	// 測試案例：舊包裹格式。
	t.Run("LegacyFormat", func(t *testing.T) {
		data := []byte(`{"data": ["c", "d"]}`)
		var arr JSONBArray
		err := arr.Scan(data)
		if err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if len(arr.Data) != 2 || arr.Data[0].(string) != "c" {
			t.Errorf("Expected [c d], got %v", arr.Data)
		}
	})
}

// TestJSONBMap tests the JSONBMap type's methods.
// TestJSONBMap 測試 JSONBMap 類型的方法。
func TestJSONBMap(t *testing.T) {
	// Test case: Empty map.
	// 測試案例：空映射。
	t.Run("Empty", func(t *testing.T) {
		var m JSONBMap
		val, err := m.Value()
		if err != nil {
			t.Fatalf("Value() error: %v", err)
		}
		if val != nil {
			t.Errorf("Expected nil, got %v", val)
		}

		// MarshalJSON should return {}.
		// MarshalJSON 應返回 {}。
		marshaled, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}
		if string(marshaled) != "{}" {
			t.Errorf("Expected {}, got %s", marshaled)
		}
	})

	// Test case: Nil data.
	// 測試案例：Nil 資料。
	t.Run("Nil", func(t *testing.T) {
		var m JSONBMap // Use value type to get non-nil pointer. / 使用值類型以取得非 nil 指標。
		err := (&m).Scan(nil)
		if err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if m.Data != nil {
			t.Errorf("Expected Data nil, got %v", m.Data)
		}
	})

	// Test case: Direct format map.
	// 測試案例：直接格式映射。
	t.Run("DirectFormat", func(t *testing.T) {
		data := []byte(`{"key": "value"}`)
		var m JSONBMap
		err := m.Scan(data)
		if err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if m.Data["key"].(string) != "value" {
			t.Errorf("Expected key:value, got %v", m.Data)
		}

		// Value should return the direct map.
		// Value 應返回直接映射。
		val, _ := m.Value()
		if string(val.([]byte)) != `{"key":"value"}` {
			t.Errorf("Expected {\"key\":\"value\"}, got %s", val)
		}

		// MarshalJSON should return direct map.
		// MarshalJSON 應返回直接映射。
		marshaled, _ := json.Marshal(m)
		if string(marshaled) != `{"key":"value"}` {
			t.Errorf("Expected {\"key\":\"value\"}, got %s", marshaled)
		}

		// UnmarshalJSON.
		// UnmarshalJSON。
		err = json.Unmarshal(data, &m)
		if err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}
		if len(m.Data) != 1 {
			t.Errorf("Expected length 1, got %d", len(m.Data))
		}
	})

	// Test case: Legacy wrapped format.
	// 測試案例：舊包裹格式。
	t.Run("LegacyFormat", func(t *testing.T) {
		data := []byte(`{"data": {"foo": "bar"}}`)
		var m JSONBMap
		err := m.Scan(data)
		if err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if m.Data["foo"].(string) != "bar" {
			t.Errorf("Expected foo:bar, got %v", m.Data)
		}
	})
}

// Additional test for JSONBValue and JSONBScan helpers.
// 額外測試 JSONBValue 和 JSONBScan 輔助函數。
func TestJSONBHelpers(t *testing.T) {
	type TestStruct struct {
		Field string `json:"field"`
	}
	model := TestStruct{Field: "test"}

	// Test JSONBValue.
	// 測試 JSONBValue。
	val, err := JSONBValue(model)
	if err != nil {
		t.Fatalf("JSONBValue error: %v", err)
	}
	expected := `{"field":"test"}`
	if string(val.([]byte)) != expected {
		t.Errorf("Expected %s, got %s", expected, val)
	}

	// Test JSONBScan.
	// 測試 JSONBScan。
	var scanned TestStruct
	err = JSONBScan(&scanned, val)
	if err != nil {
		t.Fatalf("JSONBScan error: %v", err)
	}
	if scanned.Field != "test" {
		t.Errorf("Expected field 'test', got %s", scanned.Field)
	}
}

// TestJSONBMapArray tests the JSONBMapArray type.
// TestJSONBMapArray 測試 JSONBMapArray 類型。
func TestJSONBMapArray(t *testing.T) {
	arr := JSONBMapArray{
		e2map.Map{"key1": "val1"},
		e2map.Map{"key2": "val2"},
	}

	// Value.
	// Value。
	val, err := arr.Value()
	if err != nil {
		t.Fatalf("Value error: %v", err)
	}
	expected := `[{"key1":"val1"},{"key2":"val2"}]`
	if string(val.([]byte)) != expected {
		t.Errorf("Expected %s, got %s", expected, val)
	}

	// Scan.
	// Scan。
	var scanned JSONBMapArray
	err = scanned.Scan(val)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(scanned) != 2 || scanned[0]["key1"] != "val1" {
		t.Errorf("Expected length 2 with val1, got %v", scanned)
	}
}
