package redigomock

import (
	"fmt"
	"sort"
)

const (
	_redisKey       = 0
	_redisSet       = 1
	_redisSortedSet = 2
	_redisHashSet	= 3
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
	if values[i].score == values[j].score {
		return toString(values[i].value) < toString(values[j].value)
	}
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

func (f *fakeRedis) getHashSet(key string) (map[string]interface{}, error) {
	_struct, err := f.getRedisStruct(key, _redisHashSet)
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

func (f *fakeRedis) sortedSetEnum(args []interface{}, callback func(set map[string]*scoredValue, value *scoredValue)) error {
	if len(args) < 3 {
		return fmt.Errorf("Wrong number of arguments passed")
	}
	key := toString(args[0])
	fromArg := toString(args[1])
	toArg := toString(args[2])
	set, err := f.getSortedSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return nil
	}
	values := make(scoredValueArray, 0, len(set))
	for _, value := range set {
		values = append(values, value)
	}
	sort.Sort(values)
	fromOperator, from, err := extractOperator(fromArg, gte, gt)
	if err != nil {
		return err
	}
	toOperator, to, err := extractOperator(toArg, lte, lt)
	if err != nil {
		return err
	}
	for _, v := range values {
		if fromOperator(v.score, from) && toOperator(v.score, to) {
			callback(set, v)
		}
	}
	return nil
}
