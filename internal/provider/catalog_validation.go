package provider

import (
	"fmt"
	"reflect"
	"strings"
)

func validateProviderCatalogArgs(componentName string, args any) error {
	entry, ok := providerCatalog[componentName]
	if !ok {
		return fmt.Errorf("unknown provider component %s", componentName)
	}

	value := reflect.ValueOf(args)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return fmt.Errorf("%s args must not be nil", componentName)
		}
		value = value.Elem()
	}

	for _, field := range entry.RequiredInputs {
		if isGoFieldMissing(value, field) {
			return fmt.Errorf("%s requires input %s", componentName, field)
		}
	}

	for _, group := range entry.RequiredOneOf {
		present := false
		for _, field := range group {
			if !isGoFieldMissing(value, field) {
				present = true
				break
			}
		}
		if !present {
			return fmt.Errorf("%s requires one of: %s", componentName, strings.Join(group, ", "))
		}
	}

	for _, group := range entry.MutuallyExclusive {
		present := 0
		for _, field := range group {
			if !isGoFieldMissing(value, field) {
				present++
			}
		}
		if present > 1 {
			return fmt.Errorf("%s accepts only one of: %s", componentName, strings.Join(group, ", "))
		}
	}

	if entry.ExperimentalOptInField != "" && !boolFieldIsTrue(value, entry.ExperimentalOptInField) {
		return fmt.Errorf("%s requires %s: true", componentName, entry.ExperimentalOptInField)
	}

	return nil
}

func isGoFieldMissing(value reflect.Value, pulumiName string) bool {
	field, ok := fieldByPulumiName(value, pulumiName)
	if !ok {
		return true
	}
	if field.Kind() == reflect.Pointer {
		return field.IsNil() || field.Elem().IsZero()
	}
	return field.IsZero()
}

func boolFieldIsTrue(value reflect.Value, pulumiName string) bool {
	field, ok := fieldByPulumiName(value, pulumiName)
	if !ok {
		return false
	}
	if field.Kind() == reflect.Bool {
		return field.Bool()
	}
	if field.Kind() == reflect.Pointer && !field.IsNil() && field.Elem().Kind() == reflect.Bool {
		return field.Elem().Bool()
	}
	return false
}

func fieldByPulumiName(value reflect.Value, pulumiName string) (reflect.Value, bool) {
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		structField := valueType.Field(i)
		if pulumiTagName(structField.Tag.Get("pulumi")) == pulumiName {
			return value.Field(i), true
		}
	}
	return reflect.Value{}, false
}

func pulumiTagName(tag string) string {
	if tag == "" || tag == "-" {
		return ""
	}
	if comma := strings.Index(tag, ","); comma >= 0 {
		return tag[:comma]
	}
	return tag
}
