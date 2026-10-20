package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var validate = func() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]

		if name == "" || name == "-" {
			return field.Name
		}

		return name
	})

	return v
}()

func LoadEnv() error {
	err := godotenv.Load()
	{
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("load .env: %w", err)
		}
	}

	return nil
}

func ENV(key string) string {
	return os.Getenv(key)
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	{
		if err := json.NewEncoder(w).Encode(data); err != nil {
			return
		}
	}
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, data any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()
	{
		if err := decoder.Decode(data); err != nil {
			return errors.New("invalid JSON body")
		}

		if err := decoder.Decode(new(any)); err != io.EOF {
			return errors.New("body must contain a single JSON object")
		}
	}

	return nil
}

func Validate(data any) map[string]string {
	err := validate.Struct(data)
	{
		if err == nil {
			return nil
		}
	}

	errs := make(map[string]string)

	var validationErrors validator.ValidationErrors
	{
		if !errors.As(err, &validationErrors) {
			errs["validation"] = "invalid input"
			return errs
		}
	}

	for _, field := range validationErrors {
		switch field.Tag() {
		case "required":
			errs[field.Field()] = field.Field() + " is required"
		case "email":
			errs[field.Field()] = "invalid email address"
		case "min":
			errs[field.Field()] = field.Field() + " must contain at least " + field.Param() + " characters"
		case "max":
			errs[field.Field()] = field.Field() + " must contain at most " + field.Param() + " characters"
		default:
			errs[field.Field()] = "invalid value"
		}
	}

	return errs
}
