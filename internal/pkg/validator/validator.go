package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate *validator.Validate

func Init() {
	validate = validator.New()
	registerCustomValidations(validate)
}

func ValidateStruct(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)

	for _, err := range err.(validator.ValidationErrors) {
		field := strings.ToLower(err.Field())
		errors[field] = getValidationMessage(err)
	}

	return errors
}

func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "Это поле обязательно для заполнения"
	case "email":
		return "Введите корректный email адрес"
	case "min":
		return fmt.Sprintf("Минимальная длина: %s символов", err.Param())
	case "max":
		return fmt.Sprintf("Максимальная длина: %s символов", err.Param())
	case "uuid":
		return "Неверный формат идентификатора"
	case "number":
		return "Должно быть числом"
	default:
		return fmt.Sprintf("Некорректное значение для поля %s", err.Field())
	}
}

func registerCustomValidations(v *validator.Validate) {
	v.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		uuidStr := fl.Field().String()
		_, err := uuid.Parse(uuidStr)
		return err == nil
	})

	v.RegisterValidation("number", func(fl validator.FieldLevel) bool {
		field := fl.Field()
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return true
		case reflect.String:
			str := field.String()
			if str == "" {
				return false
			}
			for _, char := range str {
				if char < '0' || char > '9' {
					return false
				}
			}
			return true
		default:
			return false
		}
	})
}
