package io

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/stephenfire/go-rtl"
)

func ConvertInputsForSerialization(executionInputs []interface{}) ([][]byte, error) {
	inputs := [][]byte{}

	for _, input := range executionInputs {
		buf := new(bytes.Buffer)

		// just get the real one
		if reflect.TypeOf(input).Kind() == reflect.Ptr {
			input = reflect.ValueOf(input).Elem().Interface()
		}

		if err := rtl.Encode(input, buf); err != nil {
			return nil, err
		}
		inputs = append(inputs, buf.Bytes())
	}

	return inputs, nil
}

func ConvertOutputsForSerialization(executionOutputs []interface{}) ([][]byte, error) {

	outputs := [][]byte{}

	for _, output := range executionOutputs {
		buf := new(bytes.Buffer)

		// just get the real one
		if reflect.TypeOf(output).Kind() == reflect.Ptr {
			output = reflect.ValueOf(output).Elem().Interface()
		}

		if err := rtl.Encode(output, buf); err != nil {
			return nil, err
		}
		outputs = append(outputs, buf.Bytes())
	}

	return outputs, nil
}

func ConvertInputsFromSerialization(handlerInfo types.HandlerInfo, executionInputs [][]byte) ([]interface{}, error) {
	inputs := []interface{}{}

	for idx, inputType := range handlerInfo.ParamTypes {
		buf := bytes.NewBuffer(executionInputs[idx])

		// Get the pointer of the type of the parameter that we target
		decodedObj := reflect.New(inputType).Elem().Addr().Interface()

		if err := rtl.Decode(buf, decodedObj); err != nil {
			return nil, err
		}

		inputs = append(inputs, reflect.ValueOf(decodedObj).Elem().Interface())
	}

	return inputs, nil
}

func ConvertOutputsFromSerialization(handlerInfo types.HandlerInfo, executionOutputs [][]byte) ([]interface{}, error) {
	output := []interface{}{}

	for idx, outputType := range handlerInfo.ReturnTypes {
		buf := bytes.NewBuffer(executionOutputs[idx])

		// Get the pointer of the type of the parameter that we target
		decodedObj := reflect.New(outputType).Elem().Addr().Interface()

		if err := rtl.Decode(buf, decodedObj); err != nil {
			return nil, err
		}

		output = append(output, reflect.ValueOf(decodedObj).Elem().Interface())
	}

	return output, nil
}

func ConvertInputsForSerializationFromValues(regularValues []interface{}) ([][]byte, error) {
	inputs := [][]byte{}

	for _, inputPointer := range regularValues {
		buf := new(bytes.Buffer)

		decodedObj := reflect.ValueOf(inputPointer).Interface()

		if err := rtl.Encode(decodedObj, buf); err != nil {
			return nil, err
		}
		inputs = append(inputs, buf.Bytes())
	}

	return inputs, nil
}

func ConvertOutputsFromSerializationToPointer(pointerValues []interface{}, executionOutputs [][]byte) ([]interface{}, error) {
	output := []interface{}{}

	for idx, outputPointers := range pointerValues {
		buf := bytes.NewBuffer(executionOutputs[idx])

		if reflect.TypeOf(outputPointers).Kind() != reflect.Ptr {
			return nil, fmt.Errorf("The output type is not a pointer")
		}

		// Get the pointer of the type of the parameter that we target
		decodedObj := reflect.New(reflect.TypeOf(outputPointers).Elem()).Interface()

		if err := rtl.Decode(buf, decodedObj); err != nil {
			return nil, err
		}

		// assign the decoded value (like `bool`) to the pointer (like `*bool`)
		reflect.ValueOf(outputPointers).Elem().Set(reflect.ValueOf(decodedObj).Elem())

		output = append(output, reflect.ValueOf(decodedObj).Elem().Interface())
	}

	return output, nil
}
