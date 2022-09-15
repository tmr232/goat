package main

import (
	"fmt"
	"github.com/pkg/errors"
)

func Reversed[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, value := range slice {
		result[len(slice)-i-1] = value
	}
	return result
}

func Map[T any, R any](slice []T, op func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = op(v)
	}
	return result
}

func MapE[T any, R any](slice []T, op func(T) (R, error)) ([]R, error) {
	result := make([]R, len(slice))
	var err error
	for i, v := range slice {
		result[i], err = op(v)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("op %v failed to map over value: %v", &op, v))
		}
	}
	return result, nil
}
