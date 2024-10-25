package tempolite

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

func convertIO(rawInput interface{}, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	if rawInput == nil {
		return nil, fmt.Errorf("nil input not supported")
	}

	rawValue := reflect.ValueOf(rawInput)

	// If desired type is a pointer
	if desiredKind == reflect.Ptr {
		return convertToPointer(rawInput, desiredType)
	}

	// Handle pointer input by dereferencing
	if rawValue.Kind() == reflect.Ptr {
		if rawValue.IsNil() {
			return nil, fmt.Errorf("nil pointer not supported")
		}
		rawValue = rawValue.Elem()
		rawInput = rawValue.Interface()
	}

	// If types match exactly after potential dereferencing, return as is
	if rawValue.Type() == desiredType {
		return rawInput, nil
	}

	// Handle named types (custom types based on basic types)
	if desiredType.PkgPath() != "" || desiredType.Name() != "" {
		return handleNamedType(rawInput, desiredType)
	}

	// Regular type conversions based on kind
	return convertBasedOnKind(rawInput, desiredType, desiredKind)
}

func handleNamedType(input interface{}, desiredType reflect.Type) (interface{}, error) {
	inputValue := reflect.ValueOf(input)
	baseKind := desiredType.Kind()

	// Create new instance of the named type
	newValue := reflect.New(desiredType).Elem()

	// If input is a named type, get its underlying value
	if inputValue.Type().Name() != "" {
		// Use the underlying value based on the kind
		switch inputValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			input = inputValue.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			input = inputValue.Uint()
		case reflect.Float32, reflect.Float64:
			input = inputValue.Float()
		case reflect.String:
			input = inputValue.String()
		case reflect.Bool:
			input = inputValue.Bool()
		case reflect.Slice:
			if inputValue.Type().Elem().Kind() == reflect.Uint8 {
				input = inputValue.Bytes()
			}
		}
	}

	switch baseKind {
	case reflect.String:
		str, err := toString(input)
		if err != nil {
			return nil, err
		}
		newValue.SetString(str)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := toInt64(input)
		if err != nil {
			return nil, err
		}
		if err := checkIntOverflow(i, baseKind); err != nil {
			return nil, err
		}
		newValue.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := toUint64(input)
		if err != nil {
			return nil, err
		}
		if err := checkUintOverflow(u, baseKind); err != nil {
			return nil, err
		}
		newValue.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := toFloat64(input)
		if err != nil {
			return nil, err
		}
		if baseKind == reflect.Float32 && (f > math.MaxFloat32 || f < -math.MaxFloat32) {
			return nil, fmt.Errorf("value %v overflows float32", f)
		}
		newValue.SetFloat(f)

	case reflect.Bool:
		b, err := toBool(input)
		if err != nil {
			return nil, err
		}
		newValue.SetBool(b)

	case reflect.Slice:
		if desiredType.Elem().Kind() == reflect.Uint8 {
			bytes, err := toBytes(input)
			if err != nil {
				return nil, err
			}
			newValue.SetBytes(bytes)
		} else {
			convertedSlice, err := convertToSlice(input, desiredType)
			if err != nil {
				return nil, err
			}
			newValue.Set(reflect.ValueOf(convertedSlice))
		}

	default:
		converted, err := convertBasedOnKind(input, desiredType, baseKind)
		if err != nil {
			return nil, err
		}
		newValue.Set(reflect.ValueOf(converted))
	}

	return newValue.Interface(), nil
}

func convertBasedOnKind(rawInput interface{}, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	switch desiredKind {
	case reflect.String:
		return toString(rawInput)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := toInt64(rawInput)
		if err != nil {
			return nil, err
		}
		if err := checkIntOverflow(i, desiredKind); err != nil {
			return nil, err
		}
		return convertIntToType(i, desiredType, desiredKind)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := toUint64(rawInput)
		if err != nil {
			return nil, err
		}
		if err := checkUintOverflow(u, desiredKind); err != nil {
			return nil, err
		}
		return convertUintToType(u, desiredType, desiredKind)

	case reflect.Float32, reflect.Float64:
		f, err := toFloat64(rawInput)
		if err != nil {
			return nil, err
		}
		return convertFloatToType(f, desiredType, desiredKind)

	case reflect.Bool:
		return toBool(rawInput)

	case reflect.Slice:
		return convertToSlice(rawInput, desiredType)

	case reflect.Array:
		return convertToArray(rawInput, desiredType)

	case reflect.Map:
		return convertToMap(rawInput, desiredType)

	case reflect.Struct:
		return convertToStruct(rawInput, desiredType)
	}

	return nil, fmt.Errorf("unsupported conversion from %T to %v", rawInput, desiredKind)
}

func convertToPointer(rawInput interface{}, desiredType reflect.Type) (interface{}, error) {
	// Create a new pointer of the desired type
	ptrValue := reflect.New(desiredType.Elem())

	// If input is already a pointer, dereference it
	inputValue := reflect.ValueOf(rawInput)
	if inputValue.Kind() == reflect.Ptr {
		if inputValue.IsNil() {
			return nil, fmt.Errorf("nil pointer not supported")
		}
		rawInput = inputValue.Elem().Interface()
	}

	// Convert the input to the element type
	converted, err := convertIO(rawInput, desiredType.Elem(), desiredType.Elem().Kind())
	if err != nil {
		return nil, fmt.Errorf("failed to convert pointer element: %w", err)
	}

	// Set the converted value
	ptrValue.Elem().Set(reflect.ValueOf(converted))
	return ptrValue.Interface(), nil
}

func toString(rawInput interface{}) (string, error) {
	switch v := rawInput.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10), nil
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10), nil
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func toInt64(rawInput interface{}) (int64, error) {
	switch v := rawInput.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint, uint8, uint16, uint32, uint64:
		result := reflect.ValueOf(v).Uint()
		if result > math.MaxInt64 {
			return 0, fmt.Errorf("uint value %v overflows int64", result)
		}
		return int64(result), nil
	case float32, float64:
		f := reflect.ValueOf(v).Float()
		if f < math.MinInt64 || f > math.MaxInt64 {
			return 0, fmt.Errorf("float value %v overflows int64", f)
		}
		return int64(f), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", rawInput)
	}
}

func toUint64(rawInput interface{}) (uint64, error) {
	switch v := rawInput.(type) {
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v).Int()
		if i < 0 {
			return 0, fmt.Errorf("negative value %v cannot be converted to uint64", i)
		}
		return uint64(i), nil
	case float32, float64:
		f := reflect.ValueOf(v).Float()
		if f < 0 || f > math.MaxUint64 {
			return 0, fmt.Errorf("float value %v cannot be converted to uint64", f)
		}
		return uint64(f), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", rawInput)
	}
}

func toFloat64(rawInput interface{}) (float64, error) {
	switch v := rawInput.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint()), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", rawInput)
	}
}

func toBool(rawInput interface{}) (bool, error) {
	switch v := rawInput.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0, nil
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", rawInput)
	}
}

func toBytes(input interface{}) ([]byte, error) {
	switch v := input.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		val := reflect.ValueOf(input)
		if val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.Uint8 {
			return val.Bytes(), nil
		}
		return nil, fmt.Errorf("cannot convert %T to []byte", input)
	}
}

func checkIntOverflow(value int64, kind reflect.Kind) error {
	switch kind {
	case reflect.Int:
		if value < math.MinInt || value > math.MaxInt {
			return fmt.Errorf("value %v overflows int", value)
		}
	case reflect.Int8:
		if value < math.MinInt8 || value > math.MaxInt8 {
			return fmt.Errorf("value %v overflows int8", value)
		}
	case reflect.Int16:
		if value < math.MinInt16 || value > math.MaxInt16 {
			return fmt.Errorf("value %v overflows int16", value)
		}
	case reflect.Int32:
		if value < math.MinInt32 || value > math.MaxInt32 {
			return fmt.Errorf("value %v overflows int32", value)
		}
	}
	return nil
}

func checkUintOverflow(value uint64, kind reflect.Kind) error {
	switch kind {
	case reflect.Uint:
		if value > math.MaxUint {
			return fmt.Errorf("value %v overflows uint", value)
		}
	case reflect.Uint8:
		if value > math.MaxUint8 {
			return fmt.Errorf("value %v overflows uint8", value)
		}
	case reflect.Uint16:
		if value > math.MaxUint16 {
			return fmt.Errorf("value %v overflows uint16", value)
		}
	case reflect.Uint32:
		if value > math.MaxUint32 {
			return fmt.Errorf("value %v overflows uint32", value)
		}
	}
	return nil
}

func convertIntToType(value int64, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	switch desiredKind {
	case reflect.Int:
		return int(value), nil
	case reflect.Int8:
		return int8(value), nil
	case reflect.Int16:
		return int16(value), nil
	case reflect.Int32:
		return int32(value), nil
	case reflect.Int64:
		return value, nil
	default:
		return nil, fmt.Errorf("cannot convert int64 to %v", desiredKind)
	}
}
func convertUintToType(value uint64, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	switch desiredKind {
	case reflect.Uint:
		return uint(value), nil
	case reflect.Uint8:
		return uint8(value), nil
	case reflect.Uint16:
		return uint16(value), nil
	case reflect.Uint32:
		return uint32(value), nil
	case reflect.Uint64:
		return value, nil
	default:
		return nil, fmt.Errorf("cannot convert uint64 to %v", desiredKind)
	}
}

func convertFloatToType(value float64, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	switch desiredKind {
	case reflect.Float32:
		return float32(value), nil
	case reflect.Float64:
		return value, nil
	default:
		return nil, fmt.Errorf("cannot convert float64 to %v", desiredKind)
	}
}

func convertToSlice(rawInput interface{}, desiredType reflect.Type) (interface{}, error) {
	inputVal := reflect.ValueOf(rawInput)

	// Special handling for string to []byte conversion
	if desiredType.Elem().Kind() == reflect.Uint8 {
		if str, ok := rawInput.(string); ok {
			return []byte(str), nil
		}
	}

	// Handle string to any slice conversion
	if str, ok := rawInput.(string); ok && desiredType.Kind() == reflect.Slice {
		runes := []rune(str)
		slice := reflect.MakeSlice(desiredType, len(runes), len(runes))
		for i, r := range runes {
			converted, err := convertIO(r, desiredType.Elem(), desiredType.Elem().Kind())
			if err != nil {
				return nil, fmt.Errorf("error converting string character at index %d: %w", i, err)
			}
			slice.Index(i).Set(reflect.ValueOf(converted))
		}
		return slice.Interface(), nil
	}

	if inputVal.Kind() != reflect.Slice && inputVal.Kind() != reflect.Array {
		return nil, fmt.Errorf("cannot convert %T to slice", rawInput)
	}

	length := inputVal.Len()
	slice := reflect.MakeSlice(desiredType, length, length)

	for i := 0; i < length; i++ {
		elem := inputVal.Index(i).Interface()
		converted, err := convertIO(elem, desiredType.Elem(), desiredType.Elem().Kind())
		if err != nil {
			return nil, fmt.Errorf("error converting slice element %d: %w", i, err)
		}
		slice.Index(i).Set(reflect.ValueOf(converted))
	}

	return slice.Interface(), nil
}

func convertToArray(rawInput interface{}, desiredType reflect.Type) (interface{}, error) {
	inputVal := reflect.ValueOf(rawInput)

	if inputVal.Kind() != reflect.Array && inputVal.Kind() != reflect.Slice {
		return nil, fmt.Errorf("cannot convert %T to array", rawInput)
	}

	arrayValue := reflect.New(desiredType).Elem()
	length := inputVal.Len()
	if length > arrayValue.Len() {
		length = arrayValue.Len()
	}

	for i := 0; i < length; i++ {
		elem := inputVal.Index(i).Interface()
		converted, err := convertIO(elem, desiredType.Elem(), desiredType.Elem().Kind())
		if err != nil {
			return nil, fmt.Errorf("error converting array element %d: %w", i, err)
		}
		arrayValue.Index(i).Set(reflect.ValueOf(converted))
	}

	return arrayValue.Interface(), nil
}

func convertToMap(rawInput interface{}, desiredType reflect.Type) (interface{}, error) {
	inputVal := reflect.ValueOf(rawInput)
	if inputVal.Kind() != reflect.Map {
		return nil, fmt.Errorf("cannot convert %T to map", rawInput)
	}

	newMap := reflect.MakeMap(desiredType)
	iter := inputVal.MapRange()

	for iter.Next() {
		key := iter.Key().Interface()
		val := iter.Value().Interface()

		convertedKey, err := convertIO(key, desiredType.Key(), desiredType.Key().Kind())
		if err != nil {
			return nil, fmt.Errorf("error converting map key: %w", err)
		}

		convertedVal, err := convertIO(val, desiredType.Elem(), desiredType.Elem().Kind())
		if err != nil {
			return nil, fmt.Errorf("error converting map value: %w", err)
		}

		newMap.SetMapIndex(reflect.ValueOf(convertedKey), reflect.ValueOf(convertedVal))
	}

	return newMap.Interface(), nil
}

func convertToStruct(rawInput interface{}, desiredType reflect.Type) (interface{}, error) {
	inputVal := reflect.ValueOf(rawInput)
	if inputVal.Kind() == reflect.Ptr {
		if inputVal.IsNil() {
			return nil, fmt.Errorf("nil pointer not supported")
		}
		inputVal = inputVal.Elem()
	}

	if inputVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("cannot convert %T to struct", rawInput)
	}

	newStruct := reflect.New(desiredType).Elem()

	for i := 0; i < desiredType.NumField(); i++ {
		field := desiredType.Field(i)
		if !field.IsExported() {
			continue
		}

		// Try to find corresponding field by name
		inputField := inputVal.FieldByName(field.Name)
		if !inputField.IsValid() {
			// Try JSON tag if direct name match fails
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "-" {
					inputField = inputVal.FieldByName(parts[0])
				}
			}
			if !inputField.IsValid() {
				continue
			}
		}

		// Convert the field value
		converted, err := convertIO(inputField.Interface(), field.Type, field.Type.Kind())
		if err != nil {
			return nil, fmt.Errorf("error converting field %s: %w", field.Name, err)
		}

		newStruct.Field(i).Set(reflect.ValueOf(converted))
	}

	return newStruct.Interface(), nil
}
