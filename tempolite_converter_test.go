package tempolite

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type BookHexID string

func (b BookHexID) String() string {
	return string(b)
}

type BookID string

func (b BookID) String() string {
	return string(b)
}

func GIDToBase64(gid []byte) string {
	return base64.StdEncoding.EncodeToString(gid)
}

type BookGID []byte

func (b BookGID) Bytes() []byte {
	return []byte(b)
}

func (b BookGID) String() string {
	return GIDToBase64(b)
}

type BookUri string

func (b BookUri) String() string {
	return string(b)
}

type BookIdentifier struct {
	HexID BookHexID // Example: "59WN2psjkt1tyaxjspN8fp"
	ID    BookID    // Example: "1QEEqeFIZktqIpPI4jSVSF"
	GID   BookGID   // Example: []byte(PML+YsMKTA+vU7gOqihgoQ==)
	URI   BookUri   // Example: "book:1QEEqeFIZktqIpPI4jSVSF"
}

func (b BookIdentifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		HexID string `json:"hexid"`
		ID    string `json:"id"`
		GID   string `json:"gid"`
		URI   string `json:"uri"`
	}{
		HexID: b.HexID.String(),
		ID:    b.ID.String(),
		GID:   base64.StdEncoding.EncodeToString(b.GID),
		URI:   b.URI.String(),
	})
}

func (b *BookIdentifier) UnmarshalJSON(data []byte) error {
	aux := struct {
		HexID string `json:"hexid"`
		ID    string `json:"id"`
		GID   string `json:"gid"`
		URI   string `json:"uri"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	b.HexID = BookHexID(aux.HexID)
	b.ID = BookID(aux.ID)
	b.URI = BookUri(aux.URI)

	gid, err := base64.StdEncoding.DecodeString(aux.GID)
	if err != nil {
		return fmt.Errorf("failed to decode GID: %v", err)
	}
	b.GID = BookGID(gid)

	return nil
}

// go test -timeout 30s -v -run ^TestConverterBookIdentifier$  .
func TestConverterBookIdentifier(t *testing.T) {
	rawInput := BookIdentifier{
		HexID: BookHexID("59WN2psjkt1tyaxjspN8fp"),
		ID:    BookID("1QEEqeFIZktqIpPI4jSVSF"),
		GID:   BookGID([]byte("PML+YsMKTA+vU7gOqihgoQ==")),
	}
	desiredType := reflect.TypeOf(rawInput)
	desiredKind := desiredType.Kind()

	convertedValue, err := convertIO(rawInput, desiredType, desiredKind)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fail()
	} else {
		scustomID := convertedValue.(BookIdentifier)
		fmt.Println("Converted book:", scustomID)
	}
}

// go test -timeout 30s -v -run ^TestConverterBookIdentifierPointer$  .
func TestConverterBookIdentifierPointer(t *testing.T) {
	rawInput := &BookIdentifier{
		HexID: BookHexID("59WN2psjkt1tyaxjspN8fp"),
		ID:    BookID("1QEEqeFIZktqIpPI4jSVSF"),
		GID:   BookGID([]byte("PML+YsMKTA+vU7gOqihgoQ==")),
	}
	desiredType := reflect.TypeOf(rawInput)
	desiredKind := desiredType.Kind()

	convertedValue, err := convertIO(rawInput, desiredType, desiredKind)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fail()
	} else {
		scustomID := convertedValue.(*BookIdentifier)
		fmt.Println("Converted book:", scustomID)
	}
}

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

type NestedStructPtr struct {
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

	nested := NestedStructPtr{
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

type CustomStruct struct {
	Title   string  `json:"title"`
	Score   float64 `json:"score"`
	Tags    []string
	Details map[string]interface{}
}

type CustomString string
type CustomSlice []string
type CustomMap map[string]int

// Interface implementation test
type Stringer interface {
	String() string
}

func (c CustomString) String() string {
	return string(c)
}

// Struct definitions
type TestStruct struct {
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	IsAdmin bool    `json:"is_admin"`
	Score   float64 `json:"score"`
}

type InnerStruct struct {
	Value    *int     `json:"value"`
	Optional string   `json:"optional"`
	Tags     []string `json:"tags"`
}

type NestedStruct struct {
	Data *InnerStruct   `json:"data"`
	List []CustomString `json:"list"`
	Map  map[string]int `json:"map"`
}

type ComplexSlice []struct {
	ID    CustomInt   `json:"id"`
	Items []*string   `json:"items"`
	Info  InnerStruct `json:"info"`
}

func TestConvertInputs(t *testing.T) {

	// Test activity function definitions
	primitiveActivity := func(ctx ActivityContext,
		str string,
		num int,
		flag bool,
		f float64) error {
		return nil
	}

	pointerActivity := func(ctx ActivityContext,
		strPtr *string,
		numPtr *int,
		flagPtr *bool) error {
		return nil
	}

	customTypesActivity := func(ctx ActivityContext,
		cs CustomString,
		ci *CustomInt,
		cf CustomFloat,
		cm CustomMap) error {
		return nil
	}

	sliceActivity := func(ctx ActivityContext,
		basicSlice []int,
		pointerSlice []*string,
		customSlice CustomSlice,
		complexSlice ComplexSlice) error {
		return nil
	}

	mapActivity := func(ctx ActivityContext,
		basicMap map[string]int,
		pointerMap map[string]*float64,
		customKeyMap map[CustomString]int,
		nestedMap map[string]map[string][]CustomInt) error {
		return nil
	}

	structActivity := func(ctx ActivityContext,
		basic TestStruct,
		nested *NestedStruct,
		withSlices struct {
			Items     []CustomInt
			MoreItems []*CustomString
		}) error {
		return nil
	}

	interfaceActivity := func(ctx ActivityContext,
		stringer CustomString, // implements Stringer
		anySlice []interface{},
		anyMap map[string]interface{}) error {
		return nil
	}

	// Test data setup
	str := "test"
	num := 42
	flag := true
	float := 3.14
	customInt := CustomInt(123)

	// Prepare nested test data
	innerStruct := &InnerStruct{
		Value:    &num,
		Optional: "optional",
		Tags:     []string{"tag1", "tag2"},
	}

	nestedStruct := &NestedStruct{
		Data: innerStruct,
		List: []CustomString{"one", "two"},
		Map:  map[string]int{"key": 1},
	}

	tests := []struct {
		name     string
		activity interface{}
		inputs   []interface{}
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Basic primitives",
			activity: primitiveActivity,
			inputs:   []interface{}{"test", 42, true, 3.14},
		},
		{
			name:     "Primitive pointers",
			activity: pointerActivity,
			inputs:   []interface{}{&str, &num, &flag},
		},
		{
			name:     "Custom types",
			activity: customTypesActivity,
			inputs: []interface{}{
				"test",                    // -> CustomString
				42,                        // -> *CustomInt
				float64(3.14),             // -> CustomFloat
				map[string]int{"test": 1}, // -> CustomMap
			},
		},
		{
			name:     "Slices basic",
			activity: sliceActivity,
			inputs: []interface{}{
				[]int{1, 2, 3},
				[]*string{&str},
				[]string{"a", "b", "c"},
				ComplexSlice{
					{
						ID:    customInt,
						Items: []*string{&str},
						Info:  *innerStruct,
					},
				},
			},
		},
		{
			name:     "Maps",
			activity: mapActivity,
			inputs: []interface{}{
				map[string]int{"a": 1, "b": 2},
				map[string]*float64{"a": &float},
				map[string]int{"test": 1}, // -> map[CustomString]int
				map[string]map[string][]CustomInt{
					"outer": {
						"inner": []CustomInt{1, 2, 3},
					},
				},
			},
		},
		{
			name:     "Structs",
			activity: structActivity,
			inputs: []interface{}{
				TestStruct{
					Name:    "test",
					Age:     42,
					IsAdmin: true,
					Score:   3.14,
				},
				nestedStruct,
				struct {
					Items     []CustomInt
					MoreItems []*CustomString
				}{
					Items:     []CustomInt{1, 2, 3},
					MoreItems: []*CustomString{},
				},
			},
		},
		{
			name:     "Interface implementations",
			activity: interfaceActivity,
			inputs: []interface{}{
				CustomString("test"),
				[]interface{}{1, "test", true},
				map[string]interface{}{
					"key":  "value",
					"num":  42,
					"bool": true,
				},
			},
		},
		{
			name:     "Type conversion errors",
			activity: primitiveActivity,
			inputs:   []interface{}{"test", map[string]int{}, true, 3.14},
			wantErr:  true,
			errMsg:   "cannot convert map to int",
		},
		{
			name:     "Slice conversion errors",
			activity: sliceActivity,
			inputs: []interface{}{
				map[string]string{},
				[]*string{&str},
				[]string{"a", "b"},
				[]interface{}{map[string]interface{}{}},
			},
			wantErr: true,
			errMsg:  "cannot convert map to slice",
		},
		{
			name:     "Map conversion errors",
			activity: mapActivity,
			inputs: []interface{}{
				[]string{"not", "a", "map"},
				map[string]*float64{},
				map[string]int{},
				map[string]interface{}{},
			},
			wantErr: true,
			errMsg:  "cannot convert slice to map",
		},
		{
			name:     "Struct conversion errors",
			activity: structActivity,
			inputs: []interface{}{
				[]string{"not", "a", "struct"},
				&NestedStruct{},
				struct {
					Items     []CustomInt
					MoreItems []*CustomString
				}{},
			},
			wantErr: true,
			errMsg:  "cannot convert slice to struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh Tempolite instance
			tp, err := New(
				context.Background(),
				NewRegistry().
					Activity(tt.activity).
					Build(),
				WithPath(":memory:"),
			)
			if err != nil {
				t.Fatalf("Failed to create Tempolite instance: %v", err)
			}
			defer tp.Close()

			// Get handler info
			activityValue, ok := tp.activities.Load(HandlerIdentity(runtime.FuncForPC(reflect.ValueOf(tt.activity).Pointer()).Name()))
			if !ok {
				t.Fatal("Activity not found in registry")
			}
			handlerInfo := HandlerInfo(activityValue.(Activity))

			// Try conversion
			outputs, err := tp.convertInputs(handlerInfo, tt.inputs)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("convertInputs() error = %v", err)
				return
			}

			// Verify outputs
			if len(outputs) != len(handlerInfo.ParamTypes) {
				t.Errorf("Got %d outputs, want %d", len(outputs), len(handlerInfo.ParamTypes))
				return
			}

			for i, out := range outputs {
				expectedType := handlerInfo.ParamTypes[i]
				actualType := reflect.TypeOf(out)
				if actualType == nil {
					t.Errorf("output[%d] is nil", i)
					continue
				}
				if !actualType.AssignableTo(expectedType) {
					t.Errorf("output[%d] type = %v, want %v", i, actualType, expectedType)
				}
			}
		})
	}
}
