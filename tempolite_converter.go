package tempolite

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

/// Why? Because i'm a bit paranoid

// convertIO converts rawInput into an interface{} matching the desiredType and desiredKind.
func convertIO(rawInput interface{}, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	if rawInput == nil {
		// Return zero value of desired type if rawInput is nil
		return reflect.Zero(desiredType).Interface(), nil
	}

	rawValue := reflect.ValueOf(rawInput)

	// If rawInput is already of the desired type, return it
	if rawValue.Type() == desiredType {
		return rawInput, nil
	}

	var convertedValue interface{}
	// Use desiredKind to guide the conversion
	switch desiredKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := toInt64(rawInput)
		if err != nil {
			return nil, err
		}
		convertedValue, err = convertIntToType(intValue, desiredType, desiredKind)
		if err != nil {
			return nil, err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := toUint64(rawInput)
		if err != nil {
			return nil, err
		}
		convertedValue, err = convertUintToType(uintValue, desiredType, desiredKind)
		if err != nil {
			return nil, err
		}
	case reflect.Float32, reflect.Float64:
		floatValue, err := toFloat64(rawInput)
		if err != nil {
			return nil, err
		}
		convertedValue, err = convertFloatToType(floatValue, desiredType, desiredKind)
		if err != nil {
			return nil, err
		}
	case reflect.String:
		strValue, err := toString(rawInput)
		if err != nil {
			return nil, err
		}
		convertedValue = strValue
	case reflect.Bool:
		boolValue, err := toBool(rawInput)
		if err != nil {
			return nil, err
		}
		convertedValue = boolValue
	case reflect.Slice:
		var err error
		convertedValue, err = convertToSlice(rawInput, desiredType, desiredKind)
		if err != nil {
			return nil, err
		}
	case reflect.Array:
		var err error
		convertedValue, err = convertToArray(rawInput, desiredType, desiredKind)
		if err != nil {
			return nil, err
		}
	case reflect.Map:
		var err error
		convertedValue, err = convertToMap(rawInput, desiredType, desiredKind)
		if err != nil {
			return nil, err
		}
	case reflect.Struct:
		var err error
		convertedValue, err = convertToStruct(rawInput, desiredType, desiredKind)
		if err != nil {
			return nil, err
		}
	case reflect.Ptr:
		var err error
		convertedValue, err = convertToPointer(rawInput, desiredType, desiredKind)
		if err != nil {
			return nil, err
		}
	case reflect.Interface:
		convertedValue = rawInput
	default:
		return nil, fmt.Errorf("unsupported desired kind: %v", desiredKind)
	}

	// Ensure the converted value is of desiredType
	convertedValueV := reflect.ValueOf(convertedValue)
	if convertedValueV.Type() != desiredType {
		if convertedValueV.Type().ConvertibleTo(desiredType) {
			convertedValueV = convertedValueV.Convert(desiredType)
			convertedValue = convertedValueV.Interface()
		} else {
			return nil, fmt.Errorf("cannot convert %T to %v", convertedValue, desiredType)
		}
	}

	return convertedValue, nil
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
	var baseValue interface{}
	switch desiredKind {
	case reflect.Int:
		if intValue < math.MinInt || intValue > math.MaxInt {
			return nil, fmt.Errorf("int64 value %v overflows int", intValue)
		}
		baseValue = int(intValue)
	case reflect.Int8:
		if intValue < math.MinInt8 || intValue > math.MaxInt8 {
			return nil, fmt.Errorf("int64 value %v overflows int8", intValue)
		}
		baseValue = int8(intValue)
	case reflect.Int16:
		if intValue < math.MinInt16 || intValue > math.MaxInt16 {
			return nil, fmt.Errorf("int64 value %v overflows int16", intValue)
		}
		baseValue = int16(intValue)
	case reflect.Int32:
		if intValue < math.MinInt32 || intValue > math.MaxInt32 {
			return nil, fmt.Errorf("int64 value %v overflows int32", intValue)
		}
		baseValue = int32(intValue)
	case reflect.Int64:
		baseValue = intValue
	default:
		return nil, fmt.Errorf("unsupported desired kind for int: %v", desiredKind)
	}

	// Convert baseValue to desiredType if necessary
	baseValueV := reflect.ValueOf(baseValue)
	if baseValueV.Type() != desiredType {
		if baseValueV.Type().ConvertibleTo(desiredType) {
			baseValueV = baseValueV.Convert(desiredType)
			baseValue = baseValueV.Interface()
		} else {
			return nil, fmt.Errorf("cannot convert %T to %v", baseValue, desiredType)
		}
	}

	return baseValue, nil
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
	var baseValue interface{}
	switch desiredKind {
	case reflect.Uint:
		if uintValue > math.MaxUint {
			return nil, fmt.Errorf("uint64 value %v overflows uint", uintValue)
		}
		baseValue = uint(uintValue)
	case reflect.Uint8:
		if uintValue > math.MaxUint8 {
			return nil, fmt.Errorf("uint64 value %v overflows uint8", uintValue)
		}
		baseValue = uint8(uintValue)
	case reflect.Uint16:
		if uintValue > math.MaxUint16 {
			return nil, fmt.Errorf("uint64 value %v overflows uint16", uintValue)
		}
		baseValue = uint16(uintValue)
	case reflect.Uint32:
		if uintValue > math.MaxUint32 {
			return nil, fmt.Errorf("uint64 value %v overflows uint32", uintValue)
		}
		baseValue = uint32(uintValue)
	case reflect.Uint64:
		baseValue = uintValue
	default:
		return nil, fmt.Errorf("unsupported desired kind for uint: %v", desiredKind)
	}

	// Convert baseValue to desiredType if necessary
	baseValueV := reflect.ValueOf(baseValue)
	if baseValueV.Type() != desiredType {
		if baseValueV.Type().ConvertibleTo(desiredType) {
			baseValueV = baseValueV.Convert(desiredType)
			baseValue = baseValueV.Interface()
		} else {
			return nil, fmt.Errorf("cannot convert %T to %v", baseValue, desiredType)
		}
	}

	return baseValue, nil
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
	var baseValue interface{}
	switch desiredKind {
	case reflect.Float32:
		if floatValue < -math.MaxFloat32 || floatValue > math.MaxFloat32 {
			return nil, fmt.Errorf("float64 value %v overflows float32", floatValue)
		}
		baseValue = float32(floatValue)
	case reflect.Float64:
		baseValue = floatValue
	default:
		return nil, fmt.Errorf("unsupported desired kind for float: %v", desiredKind)
	}

	// Convert baseValue to desiredType if necessary
	baseValueV := reflect.ValueOf(baseValue)
	if baseValueV.Type() != desiredType {
		if baseValueV.Type().ConvertibleTo(desiredType) {
			baseValueV = baseValueV.Convert(desiredType)
			baseValue = baseValueV.Interface()
		} else {
			return nil, fmt.Errorf("cannot convert %T to %v", baseValue, desiredType)
		}
	}

	return baseValue, nil
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
		convertedElem, err := convertIO(rawElem, elemType, elemKind)
		if err != nil {
			return nil, fmt.Errorf("cannot convert element %d: %v", i, err)
		}

		sliceElemValue := reflect.ValueOf(convertedElem)
		if sliceElemValue.Type() != elemType {
			if sliceElemValue.Type().ConvertibleTo(elemType) {
				sliceElemValue = sliceElemValue.Convert(elemType)
			} else {
				return nil, fmt.Errorf("cannot convert element %d to %v", i, elemType)
			}
		}

		sliceValue.Index(i).Set(sliceElemValue)
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

		convertedKey, err := convertIO(rawKey.Interface(), keyType, keyKind)
		if err != nil {
			return nil, fmt.Errorf("cannot convert map key %v: %v", rawKey, err)
		}

		convertedElem, err := convertIO(rawElem, elemType, elemKind)
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

		convertedValue, err := convertIO(rawFieldValue, field.Type, field.Type.Kind())
		if err != nil {
			return nil, fmt.Errorf("cannot convert field %s: %v", fieldName, err)
		}

		convertedValueV := reflect.ValueOf(convertedValue)

		if convertedValueV.Type().AssignableTo(fieldValue.Type()) {
			fieldValue.Set(convertedValueV)
		} else if convertedValueV.Type().ConvertibleTo(fieldValue.Type()) {
			fieldValue.Set(convertedValueV.Convert(fieldValue.Type()))
		} else {
			return nil, fmt.Errorf("cannot set field %s: value of type %v is not assignable to type %v", fieldName, convertedValueV.Type(), fieldValue.Type())
		}
	}

	return structValue.Interface(), nil
}

func convertToPointer(rawInput interface{}, desiredType reflect.Type, desiredKind reflect.Kind) (interface{}, error) {
	elemType := desiredType.Elem()
	elemKind := elemType.Kind()
	convertedValue, err := convertIO(rawInput, elemType, elemKind)
	if err != nil {
		return nil, err
	}
	ptrValue := reflect.New(elemType)
	convertedValueV := reflect.ValueOf(convertedValue)
	if convertedValueV.Type().AssignableTo(elemType) {
		ptrValue.Elem().Set(convertedValueV)
	} else if convertedValueV.Type().ConvertibleTo(elemType) {
		ptrValue.Elem().Set(convertedValueV.Convert(elemType))
	} else {
		return nil, fmt.Errorf("cannot convert %T to %v", convertedValue, elemType)
	}
	return ptrValue.Interface(), nil
}
