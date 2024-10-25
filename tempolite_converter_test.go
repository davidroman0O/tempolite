package tempolite

import (
	"fmt"
	"reflect"
	"testing"
)

// go test -timeout 30s -v -run ^TestConverterCustomType$  .
func TestConverterCustomType(t *testing.T) {
	rawInput := "some_id"
	desiredType := reflect.TypeOf(CustomId(""))
	desiredKind := desiredType.Kind()

	convertedValue, err := convertIO(rawInput, desiredType, desiredKind)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fail()
	} else {
		scustomID := convertedValue.(CustomId)
		fmt.Println("Converted scustomID:", scustomID)
	}
}

// go test -timeout 30s -v -run ^TestConverterArrayByteType$  .
func TestConverterArrayByteType(t *testing.T) {
	rawInput := []byte{0x01, 0x02, 0x03}
	desiredType := reflect.TypeOf([]byte{})
	desiredKind := desiredType.Kind()

	convertedValue, err := convertIO(rawInput, desiredType, desiredKind)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fail()
	} else {
		customData := convertedValue.([]byte)
		fmt.Println("Converted customData:", customData)
	}
}

// go test -timeout 30s -v -run ^TestConverterCustomByteType$  .
func TestConverterCustomByteType(t *testing.T) {
	rawInput := CustomBID([]byte{0x01, 0x02, 0x03})
	desiredType := reflect.TypeOf(CustomBID([]byte{}))
	desiredKind := desiredType.Kind()

	convertedValue, err := convertIO(rawInput, desiredType, desiredKind)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fail()
	} else {
		customID := convertedValue.(CustomBID)
		fmt.Println("Converted customID:", customID)
	}
}

// go test -timeout 30s -v -run ^TestConverterCustomStructType$  .
func TestConverterCustomStructType(t *testing.T) {
	rawInput := WithinStruct{
		Val1:    "some_val",
		Val2:    123,
		Custom:  CustomId("some_id"),
		CustomB: CustomBID([]byte{0x01, 0x02, 0x03}),
	}
	desiredType := reflect.TypeOf(WithinStruct{})
	desiredKind := desiredType.Kind()

	convertedValue, err := convertIO(rawInput, desiredType, desiredKind)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fail()
	} else {
		customStruct := convertedValue.(WithinStruct)
		fmt.Println("Converted customStruct:", customStruct)
	}
}

// Custom types for testing
type CustomId string

func (s CustomId) String() string { return string(s) }

type CustomBID []byte

func (s CustomBID) Bytes() []byte { return []byte(s) }

type CustomInt int
type CustomUint uint
type CustomFloat float64
type CustomBool bool

type WithinStruct struct {
	Val1    string
	Val2    int
	Custom  CustomId
	CustomB CustomBID
}

type NestedStruct struct {
	Basic WithinStruct
	Ptr   *WithinStruct
}

// Basic type conversions
func TestBasicTypeConversions(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		targetType  interface{}
		expectError bool
	}{
		// String conversions
		{"string to string", "test", "", false},
		{"int to string", 123, "", false},
		{"float to string", 123.45, "", false},
		{"bool to string", true, "", false},

		// Integer conversions
		{"int to int", 123, int(0), false},
		{"string to int", "123", int(0), false},
		{"float to int", 123.0, int(0), false},
		{"int8 to int16", int8(100), int16(0), false},
		{"int to int64", 123, int64(0), false},

		// Unsigned integer conversions
		{"uint to uint", uint(123), uint(0), false},
		{"int to uint", 123, uint(0), false},
		{"string to uint", "123", uint(0), false},

		// Float conversions
		{"float32 to float64", float32(123.45), float64(0), false},
		{"int to float64", 123, float64(0), false},
		{"string to float64", "123.45", float64(0), false},

		// Boolean conversions
		{"bool to bool", true, false, false},
		{"string to bool", "true", false, false},
		{"int to bool", 1, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetType := reflect.TypeOf(tt.targetType)
			result, err := convertIO(tt.input, targetType, targetType.Kind())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Got nil result")
			}
		})
	}
}

// Custom type conversions
func TestCustomTypeConversions(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		targetType  interface{}
		expectError bool
	}{
		{"string to CustomId", "test", CustomId(""), false},
		{"CustomId to string", CustomId("test"), "", false},
		{"[]byte to CustomBID", []byte{1, 2, 3}, CustomBID{}, false},
		{"CustomBID to []byte", CustomBID{1, 2, 3}, []byte{}, false},
		{"int to CustomInt", 123, CustomInt(0), false},
		{"CustomInt to int", CustomInt(123), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetType := reflect.TypeOf(tt.targetType)
			result, err := convertIO(tt.input, targetType, targetType.Kind())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Got nil result")
			}
		})
	}
}

// Pointer conversions
func TestPointerConversions(t *testing.T) {
	str := "test"
	num := 123
	customID := CustomId("test")

	tests := []struct {
		name        string
		input       interface{}
		targetType  interface{}
		expectError bool
	}{
		{"value to pointer", str, &str, false},
		{"pointer to value", &str, "", false},
		{"pointer to pointer", &num, &num, false},
		{"custom type to pointer", customID, &customID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetType := reflect.TypeOf(tt.targetType)
			result, err := convertIO(tt.input, targetType, targetType.Kind())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Got nil result")
			}
		})
	}
}

// Slice conversions
func TestSliceConversions(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		targetType  interface{}
		expectError bool
	}{
		{"[]int to []int", []int{1, 2, 3}, []int{}, false},
		{"[]int to []int64", []int{1, 2, 3}, []int64{}, false},
		{"[]byte to []byte", []byte{1, 2, 3}, []byte{}, false},
		{"string to []byte", "test", []byte{}, false},
		{"[]interface{} to []string", []interface{}{"a", "b"}, []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetType := reflect.TypeOf(tt.targetType)
			result, err := convertIO(tt.input, targetType, targetType.Kind())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Got nil result")
			}
		})
	}
}

// Map conversions
func TestMapConversions(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		targetType  interface{}
		expectError bool
	}{
		{"map[string]int to same", map[string]int{"a": 1}, map[string]int{}, false},
		{"map[string]interface{} to map[string]int",
			map[string]interface{}{"a": 1}, map[string]int{}, false},
		{"map with custom key",
			map[CustomId]int{CustomId("test"): 1}, map[CustomId]int{}, false},
		{"map with custom value",
			map[string]CustomId{"test": CustomId("value")}, map[string]CustomId{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetType := reflect.TypeOf(tt.targetType)
			result, err := convertIO(tt.input, targetType, targetType.Kind())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Got nil result")
			}
		})
	}
}

// Struct conversions
func TestStructConversions(t *testing.T) {
	basic := WithinStruct{
		Val1:    "test",
		Val2:    123,
		Custom:  CustomId("custom"),
		CustomB: CustomBID{1, 2, 3},
	}

	nested := NestedStruct{
		Basic: basic,
		Ptr:   &basic,
	}

	tests := []struct {
		name        string
		input       interface{}
		targetType  interface{}
		expectError bool
	}{
		{"basic struct", basic, WithinStruct{}, false},
		{"nested struct", nested, NestedStruct{}, false},
		{"struct pointer", &basic, WithinStruct{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetType := reflect.TypeOf(tt.targetType)
			result, err := convertIO(tt.input, targetType, targetType.Kind())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Got nil result")
			}
		})
	}
}

// Error cases
func TestErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		targetType  interface{}
		expectError bool
	}{
		{"nil input", nil, "", true},
		{"overflow int8", 1000, int8(0), true},
		{"negative to uint", -1, uint(0), true},
		{"invalid string to int", "not a number", 0, true},
		{"slice to int", []int{1, 2, 3}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetType := reflect.TypeOf(tt.targetType)
			_, err := convertIO(tt.input, targetType, targetType.Kind())

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestMapToStructConversion(t *testing.T) {
	type TestStruct struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Balance float64 `json:"balance"`
		Active  bool    `json:"active"`
	}

	tests := []struct {
		name        string
		input       map[string]interface{}
		want        TestStruct
		expectError bool
	}{
		{
			name: "basic conversion",
			input: map[string]interface{}{
				"name":    "John Doe",
				"age":     30,
				"balance": 100.50,
				"active":  true,
			},
			want: TestStruct{
				Name:    "John Doe",
				Age:     30,
				Balance: 100.50,
				Active:  true,
			},
			expectError: false,
		},
		{
			name: "partial fields",
			input: map[string]interface{}{
				"name": "Jane Doe",
				"age":  25,
			},
			want: TestStruct{
				Name: "Jane Doe",
				Age:  25,
			},
			expectError: false,
		},
		{
			name: "with type conversion",
			input: map[string]interface{}{
				"name":    "Bob",
				"age":     "40", // string that needs to be converted to int
				"balance": 200,  // int that needs to be converted to float64
				"active":  1,    // int that needs to be converted to bool
			},
			want: TestStruct{
				Name:    "Bob",
				Age:     40,
				Balance: 200.0,
				Active:  true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			desiredType := reflect.TypeOf(result)
			converted, err := convertIO(tt.input, desiredType, desiredType.Kind())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			result = converted.(TestStruct)
			if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("Got %+v, want %+v", result, tt.want)
			}
		})
	}
}
