package main

import (
	"errors"
	"fmt"
	"reflect"
)

type JsonValidator struct {
}

func NewValidator() *JsonValidator {
	return &JsonValidator{}
}

func (val *JsonValidator) CheckExpectedResponseFormat(res interface{}, expect interface{}) error {
	return val.checkExpectedResponseFormatRec("root", res, expect)
}

// Recursively goes through the response and validates the type of each property.
// It also checks for missing properties (that are present in 'expect' but missing in the response).
// supported types:
// string: not empty string
// string_empty: empty string
// number: number not 0
// number_zero: number which can be 0
// bool: any boolean
// null: JSON null
// object types must be defined using the above types for their child properties
// TODO for useful SLP response validation this needs "oneOf" from JSON schema: "one of these n keys must be defined as follows" and conditionals
func (val *JsonValidator) checkExpectedResponseFormatRec(key string, res interface{}, expect interface{}) error {
	switch res.(type) {
	case string:
		if expect == "string" && len(res.(string)) != 0 {
			break
		} else if expect == "string_empty" {
			break
		}
		return errors.New(fmt.Sprintf("'%s' string prop does not meet expected type '%s' %v", key, expect, res))
	case bool:
		if expect != "bool" {
			return errors.New(fmt.Sprintf("'%s' bool prop does not meet expected type '%s' %v", key, expect, res))
		}
	case nil:
		if expect != "null" {
			return errors.New(fmt.Sprintf("'%s' null prop does not meet expected type '%s' %v", key, expect, res))
		}
	case float64:
		if expect == "number" && res.(float64) != 0.0 {
			break
		} else if expect == "number_zero" {
			break
		}
		return errors.New(fmt.Sprintf("'%s' number prop does not meet expected type '%s' %v", key, expect, res))
	case []interface{}:
		//fmt.Printf("%s %v is a slice of interface \n", key, res)
		for _, v := range res.([]interface{}) { // recursively go through array
			// 'expect' already is the schema for each array element
			if err := val.checkExpectedResponseFormatRec(key, v, expect); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		//fmt.Printf("%s %v is a map \n", key, value)
		if err := val.checkRequiredPropertiesExist(key, res, expect); err != nil {
			return err
		}

		for k, v := range res.(map[string]interface{}) { // recursively go through map
			expectSub, exists := expect.(D)[k]
			if !exists {
				return errors.New(fmt.Sprintf("unknown interface{} prop '%s' found under '%s'", k, key))
			}
			if err := val.checkExpectedResponseFormatRec(k, v, val.getFirstObjectInArray(expectSub)); err != nil {
				return err
			}
		}
	case D: // D is defined as string map but type assertion treats it as a new type
		//fmt.Printf("%s %v is a map \n", key, value)
		if err := val.checkRequiredPropertiesExist(key, res, expect); err != nil {
			return err
		}

		for k, v := range res.(D) { //recursively go through map
			expectSub, exists := expect.(D)[k]
			if !exists {
				return errors.New(fmt.Sprintf("unknown D prop '%s' found under '%s'", k, key))
			}
			if err := val.checkExpectedResponseFormatRec(k, v, val.getFirstObjectInArray(expectSub)); err != nil {
				return err
			}
		}
	default:
		fmt.Printf("%s %v is unknown\n", key, res)
	}

	return nil
}

func (val *JsonValidator) checkRequiredPropertiesExist(key string, res interface{}, expect interface{}) error {
	switch expect.(type) {
	//case interface{}:
	//return errors.New(fmt.Sprintf("unexpected interface{} value at '%s' %v\n", key, expect))
	case map[string]interface{}:
		for k, _ := range expect.(map[string]interface{}) {
			_, exists := res.(map[string]interface{})[k]
			if !exists {
				//return errors.New(fmt.Sprintf("missing required property %s on interface{} %v\n", key, expect))
				break
			}
		}
	case D:
		for k, _ := range expect.(D) {
			switch res.(type) {
			case map[string]interface{}:
				_, exists := res.(map[string]interface{})[k]
				if !exists {
					//return errors.New(fmt.Sprintf("missing required property %s on D %v\n", key, expect))
					break
				}
			case D:
				_, exists := res.(D)[k]
				if !exists {
					//return errors.New(fmt.Sprintf("missing required property %s on D %v\n", key, expect))
					break
				}
			}
		}
	default:
		fmt.Sprintf("unable to check for expected props on '%s' %v\n", key, expect)
	}

	return nil
}

func (val *JsonValidator) getFirstObjectInArray(arrayOrObject interface{}) interface{} {
	ref := reflect.TypeOf(arrayOrObject)
	switch ref.Kind() {
	case reflect.Slice:
		return arrayOrObject.([]D)[0]
	}

	// returns D or string/float64/bool
	return arrayOrObject
}
