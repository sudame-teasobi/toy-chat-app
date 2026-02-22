package env

import (
	"fmt"
	"os"
)

type env struct {
	value        string
	defaultValue string
}

func (e *env) WithDefault(defaultValue string) *env {
	e.defaultValue = defaultValue
	return e
}

func (e *env) SafeValue() (string, error) {
	if e.value != "" {
		return e.value, nil
	}

	if e.defaultValue != "" {
		return e.defaultValue, nil
	}

	return "", fmt.Errorf("failed to get env and default value is not set")

}

func (e *env) Value() string {
	value, err := e.SafeValue()
	if err != nil {
		panic(fmt.Sprintf("failed to read environment: %s", err.Error()))
	}
	return value
}

func GetEnv(key string) *env {
	return &env{
		value: os.Getenv(key),
	}
}
