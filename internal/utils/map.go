package utils

func MapSliceE[T any, U any](data []T, mapper func(T) (U, error)) ([]U, error) {
	result := make([]U, 0, len(data))
	for _, item := range data {
		mapped, err := mapper(item)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}
	return result, nil
}

func MapSlice[T any, U any](data []T, mapper func(T) U) []U {
	result := make([]U, 0, len(data))
	for _, item := range data {
		result = append(result, mapper(item))
	}
	return result
}

func MethodMapperE[T any, U any](method func(*T) (U, error)) func(*T) (U, error) {
	return func(t *T) (U, error) {
		return method(t)
	}
}

func MethodMapper[T any, U any](method func(*T) U) func(*T) U {
	return func(t *T) U {
		return method(t)
	}
}
