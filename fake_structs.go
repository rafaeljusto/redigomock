package redigomock

import "fmt"

const (
	_redisKey       = 0
	_redisSet       = 1
	_redisSortedSet = 2
)

type scoredValue struct {
	value interface{}
	score int64
}

type scoredValueArray []*scoredValue

func (values scoredValueArray) Len() int {
	return len(values)
}

func (values scoredValueArray) Less(i, j int) bool {
	return values[i].score < values[j].score
}

func (values scoredValueArray) Swap(i, j int) {
	values[i], values[j] = values[j], values[i]
}

func (f *fakeRedis) getRedisStruct(key string, desiredType int) (interface{}, error) {
	if container, exists := f.keys[key]; exists {
		if container.redisType != desiredType {
			return nil, fmt.Errorf("Wrong type of struct. Desired '%d'. Got '%d'", desiredType, container.redisType)
		}
		return container.redisStruct, nil
	}
	return nil, nil
}

func (f *fakeRedis) getSet(key string) (map[string]interface{}, error) {
	_struct, err := f.getRedisStruct(key, _redisSet)
	if err != nil {
		return nil, err
	}
	if _struct == nil {
		return nil, nil
	}
	result := _struct.(map[string]interface{})
	return result, nil
}

func (f *fakeRedis) getSortedSet(key string) (map[string]*scoredValue, error) {
	_struct, err := f.getRedisStruct(key, _redisSortedSet)
	if err != nil {
		return nil, err
	}
	if _struct == nil {
		return nil, nil
	}
	result := _struct.(map[string]*scoredValue)
	return result, nil
}
