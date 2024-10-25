package tempolite

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/stephenfire/go-rtl"
)

func (tp *Tempolite) convertInputsForSerialization(handlerInfo HandlerInfo, executionInputs []interface{}) ([][]byte, error) {
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

func (tp *Tempolite) convertOutputsForSerialization(handlerInfo HandlerInfo, executionOutputs []interface{}) ([][]byte, error) {

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

func (tp *Tempolite) convertInputsFromSerialization(handlerInfo HandlerInfo, executionInputs [][]byte) ([]interface{}, error) {
	inputs := []interface{}{}

	for idx, inputType := range handlerInfo.ParamTypes {
		buf := bytes.NewBuffer(executionInputs[idx])

		// Get the pointer of the type of the parameter that we target
		decodedObj := reflect.New(inputType).Elem().Addr().Interface()

		// fmt.Println("convertInputsFromSerialization decodedObj before", decodedObj)
		fmt.Println("convertInputsFromSerialization serialized input", executionInputs[idx])

		if err := rtl.Decode(buf, decodedObj); err != nil {
			return nil, err
		}

		fmt.Println("convertInputsFromSerialization decodedObj", reflect.ValueOf(decodedObj).Elem().Interface())

		inputs = append(inputs, reflect.ValueOf(decodedObj).Elem().Interface())
	}

	return inputs, nil
}

func (tp *Tempolite) convertOutputsFromSerialization(handlerInfo HandlerInfo, executionOutputs [][]byte) ([]interface{}, error) {
	output := []interface{}{}

	for idx, outputType := range handlerInfo.ReturnTypes {
		buf := bytes.NewBuffer(executionOutputs[idx])

		// Get the pointer of the type of the parameter that we target
		decodedObj := reflect.New(outputType).Elem().Addr().Interface()

		// fmt.Println("convertOutputsFromSerialization decodedObj before", decodedObj)
		fmt.Println("convertOutputsFromSerialization serialized output", executionOutputs[idx])

		if err := rtl.Decode(buf, decodedObj); err != nil {
			return nil, err
		}
		fmt.Println("convertOutputsFromSerialization decodedObj", reflect.ValueOf(decodedObj).Elem().Interface())

		output = append(output, reflect.ValueOf(decodedObj).Elem().Interface())
	}

	return output, nil
}

func (tp *Tempolite) convertInputsForSerializationFromValues(regularValues []interface{}) ([][]byte, error) {
	inputs := [][]byte{}

	for _, inputPointer := range regularValues {
		buf := new(bytes.Buffer)

		fmt.Printf("convertInputsForSerializationFromValues inputPointer: %v\n", inputPointer)

		decodedObj := reflect.ValueOf(inputPointer).Interface()

		if err := rtl.Encode(decodedObj, buf); err != nil {
			return nil, err
		}
		inputs = append(inputs, buf.Bytes())
	}

	return inputs, nil
}

func (tp *Tempolite) convertOutputsFromSerializationToPointer(pointerValues []interface{}, executionOutputs [][]byte) ([]interface{}, error) {
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
