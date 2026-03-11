package redis

import "context"

// RedisMethods - методы, которые должны быть у реализации ZSET
type RedisMethods interface {

	// ZRangeByScore получает элементы из сортированного множества с баллами в интервале [min, max]
	ZRangeByScore(ctx context.Context, key string, min, max int64) ([]string, error)

	// ZRem удаляет элемент из сортированного множества
	ZRem(ctx context.Context, key string, members ...interface{}) error
}
