package tempolite

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

/// Why? Because i'm a bit paranoid

// convertInput converts rawInput into an interface{} matching the desiredType and desiredKind.
func convertInput(rawInput interface{}, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	if rawInput == nil {
		// Return zero value of desired type if rawInput is nil
		return reflect.Zero(desiredType).Interface(), nil
	}

	rawValue := reflect.ValueOf(rawInput)
	// rawKind := rawValue.Kind()

	// If rawInput is already of the desired type, return it
	if rawValue.Type() == desiredType {
		return rawInput, nil
	}

	// Use desiredKind to guide the conversion
	switch desiredKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := toInt64(rawInput)
		if err != nil {
			return nil, err
		}
		return convertIntToType(intValue, desiredType, desiredKind)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := toUint64(rawInput)
		if err != nil {
			return nil, err
		}
		return convertUintToType(uintValue, desiredType, desiredKind)
	case reflect.Float32, reflect.Float64:
		floatValue, err := toFloat64(rawInput)
		if err != nil {
			return nil, err
		}
		return convertFloatToType(floatValue, desiredType, desiredKind)
	case reflect.String:
		strValue, err := toString(rawInput)
		if err != nil {
			return nil, err
		}
		return strValue, nil
	case reflect.Bool:
		boolValue, err := toBool(rawInput)
		if err != nil {
			return nil, err
		}
		return boolValue, nil
	case reflect.Slice:
		return convertToSlice(rawInput, desiredType, desiredKind)
	case reflect.Array:
		return convertToArray(rawInput, desiredType, desiredKind)
	case reflect.Map:
		return convertToMap(rawInput, desiredType, desiredKind)
	case reflect.Struct:
		return convertToStruct(rawInput, desiredType, desiredKind)
	case reflect.Ptr:
		return convertToPointer(rawInput, desiredType, desiredKind)
	case reflect.Interface:
		return rawInput, nil
	default:
		return nil, fmt.Errorf("unsupported desired kind: %v", desiredKind)
	}
}

// Helper functions for basic types

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
		uValue, err := toUint64(v)
		if err != nil {
			return 0, err
		}
		if uValue > math.MaxInt64 {
			return 0, fmt.Errorf("uint value %v overflows int64", uValue)
		}
		return int64(uValue), nil
	case float32, float64:
		fValue, err := toFloat64(v)
		if err != nil {
			return 0, err
		}
		if fValue < math.MinInt64 || fValue > math.MaxInt64 {
			return 0, fmt.Errorf("float value %v overflows int64", fValue)
		}
		return int64(fValue), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", rawInput)
	}
}

func convertIntToType(intValue int64, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	switch desiredKind {
	case reflect.Int:
		if intValue < math.MinInt || intValue > math.MaxInt {
			return nil, fmt.Errorf("int64 value %v overflows int", intValue)
		}
		return int(intValue), nil
	case reflect.Int8:
		if intValue < math.MinInt8 || intValue > math.MaxInt8 {
			return nil, fmt.Errorf("int64 value %v overflows int8", intValue)
		}
		return int8(intValue), nil
	case reflect.Int16:
		if intValue < math.MinInt16 || intValue > math.MaxInt16 {
			return nil, fmt.Errorf("int64 value %v overflows int16", intValue)
		}
		return int16(intValue), nil
	case reflect.Int32:
		if intValue < math.MinInt32 || intValue > math.MaxInt32 {
			return nil, fmt.Errorf("int64 value %v overflows int32", intValue)
		}
		return int32(intValue), nil
	case reflect.Int64:
		return intValue, nil
	default:
		return nil, fmt.Errorf("unsupported desired kind for int: %v", desiredKind)
	}
}

func toUint64(rawInput interface{}) (uint64, error) {
	switch v := rawInput.(type) {
	case int, int8, int16, int32, int64:
		iValue, err := toInt64(v)
		if err != nil {
			return 0, err
		}
		if iValue < 0 {
			return 0, fmt.Errorf("negative int value %v cannot be converted to uint64", iValue)
		}
		return uint64(iValue), nil
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
	case float32, float64:
		fValue, err := toFloat64(v)
		if err != nil {
			return 0, err
		}
		if fValue < 0 || fValue > math.MaxUint64 {
			return 0, fmt.Errorf("float value %v cannot be converted to uint64", fValue)
		}
		return uint64(fValue), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", rawInput)
	}
}

func convertUintToType(uintValue uint64, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	switch desiredKind {
	case reflect.Uint:
		if uintValue > math.MaxUint {
			return nil, fmt.Errorf("uint64 value %v overflows uint", uintValue)
		}
		return uint(uintValue), nil
	case reflect.Uint8:
		if uintValue > math.MaxUint8 {
			return nil, fmt.Errorf("uint64 value %v overflows uint8", uintValue)
		}
		return uint8(uintValue), nil
	case reflect.Uint16:
		if uintValue > math.MaxUint16 {
			return nil, fmt.Errorf("uint64 value %v overflows uint16", uintValue)
		}
		return uint16(uintValue), nil
	case reflect.Uint32:
		if uintValue > math.MaxUint32 {
			return nil, fmt.Errorf("uint64 value %v overflows uint32", uintValue)
		}
		return uint32(uintValue), nil
	case reflect.Uint64:
		return uintValue, nil
	default:
		return nil, fmt.Errorf("unsupported desired kind for uint: %v", desiredKind)
	}
}

func toFloat64(rawInput interface{}) (float64, error) {
	switch v := rawInput.(type) {
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint()), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", rawInput)
	}
}

func convertFloatToType(floatValue float64, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	switch desiredKind {
	case reflect.Float32:
		if floatValue < -math.MaxFloat32 || floatValue > math.MaxFloat32 {
			return nil, fmt.Errorf("float64 value %v overflows float32", floatValue)
		}
		return float32(floatValue), nil
	case reflect.Float64:
		return floatValue, nil
	default:
		return nil, fmt.Errorf("unsupported desired kind for float: %v", desiredKind)
	}
}

func toString(rawInput interface{}) (string, error) {
	switch v := rawInput.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", rawInput), nil
	}
}

func toBool(rawInput interface{}) (bool, error) {
	switch v := rawInput.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", rawInput)
	}
}

// Helper functions for composite types

func convertToSlice(rawInput interface{}, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	rawValue := reflect.ValueOf(rawInput)
	if rawValue.Kind() != reflect.Slice && rawValue.Kind() != reflect.Array {
		return nil, fmt.Errorf("cannot convert %T to slice", rawInput)
	}

	length := rawValue.Len()
	sliceValue := reflect.MakeSlice(desiredType, length, length)
	elemType := desiredType.Elem()
	elemKind := elemType.Kind()

	for i := 0; i < length; i++ {
		rawElem := rawValue.Index(i).Interface()
		convertedElem, err := convertInput(rawElem, elemType, elemKind)
		if err != nil {
			return nil, fmt.Errorf("cannot convert element %d: %v", i, err)
		}
		sliceValue.Index(i).Set(reflect.ValueOf(convertedElem))
	}

	return sliceValue.Interface(), nil
}

func convertToArray(rawInput interface{}, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	// Similar to convertToSlice, but for arrays
	return convertToSlice(rawInput, desiredType, desiredKind)
}

func convertToMap(rawInput interface{}, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	rawValue := reflect.ValueOf(rawInput)
	if rawValue.Kind() != reflect.Map {
		return nil, fmt.Errorf("cannot convert %T to map", rawInput)
	}

	keyType := desiredType.Key()
	elemType := desiredType.Elem()
	keyKind := keyType.Kind()
	elemKind := elemType.Kind()
	mapValue := reflect.MakeMap(desiredType)

	for _, rawKey := range rawValue.MapKeys() {
		rawElem := rawValue.MapIndex(rawKey).Interface()

		convertedKey, err := convertInput(rawKey.Interface(), keyType, keyKind)
		if err != nil {
			return nil, fmt.Errorf("cannot convert map key %v: %v", rawKey, err)
		}

		convertedElem, err := convertInput(rawElem, elemType, elemKind)
		if err != nil {
			return nil, fmt.Errorf("cannot convert map value %v: %v", rawElem, err)
		}

		mapValue.SetMapIndex(reflect.ValueOf(convertedKey), reflect.ValueOf(convertedElem))
	}

	return mapValue.Interface(), nil
}

func convertToStruct(rawInput interface{}, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	if rawInput == nil {
		return reflect.Zero(desiredType).Interface(), nil
	}

	// If rawInput is a map, map its keys to struct fields
	var rawMap map[string]interface{}

	if m, ok := rawInput.(map[string]interface{}); ok {
		rawMap = m
	} else {
		// Attempt to convert rawInput to map[string]interface{}
		if rawValue := reflect.ValueOf(rawInput); rawValue.Kind() == reflect.Map {
			rawMap = make(map[string]interface{})
			for _, key := range rawValue.MapKeys() {
				strKey, ok := key.Interface().(string)
				if !ok {
					return nil, fmt.Errorf("map key %v is not a string", key)
				}
				rawMap[strKey] = rawValue.MapIndex(key).Interface()
			}
		} else {
			return nil, fmt.Errorf("cannot convert %T to struct", rawInput)
		}
	}

	structValue := reflect.New(desiredType).Elem()
	for i := 0; i < desiredType.NumField(); i++ {
		field := desiredType.Field(i)
		fieldValue := structValue.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		// Use JSON tag if present
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			jsonFieldName := strings.Split(jsonTag, ",")[0]
			if jsonFieldName != "" && jsonFieldName != "-" {
				fieldName = jsonFieldName
			}
		}

		rawFieldValue, exists := rawMap[fieldName]
		if !exists {
			continue
		}

		convertedValue, err := convertInput(rawFieldValue, field.Type, field.Type.Kind())
		if err != nil {
			return nil, fmt.Errorf("cannot convert field %s: %v", fieldName, err)
		}
		fieldValue.Set(reflect.ValueOf(convertedValue))
	}

	return structValue.Interface(), nil
}

func convertToPointer(rawInput interface{}, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	elemType := desiredType.Elem()
	elemKind := elemType.Kind()
	convertedValue, err := convertInput(rawInput, elemType, elemKind)
	if err != nil {
		return nil, err
	}
	ptrValue := reflect.New(elemType)
	ptrValue.Elem().Set(reflect.ValueOf(convertedValue))
	return ptrValue.Interface(), nil
}
