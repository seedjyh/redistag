// tag
package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"reflect"
	"strconv"
	"strings"
)

// 查找raw里首个单引号括起来的内容
func LookUpSingleQuote(raw string) string {
	firstIndex := strings.Index(raw, "'")
	if firstIndex < 0 {
		return ""
	}
	offset := strings.Index(raw[firstIndex+1:], "'")
	if offset < 0 {
		return ""
	}
	return raw[firstIndex+1 : firstIndex+1+offset]
}

func HMSet(redisClient redis.Cmdable, key string, v interface{}) error {
	valueMap := make(map[string]interface{})
	typeElements := reflect.TypeOf(v).Elem()
	valueElements := reflect.ValueOf(v).Elem()
	for i := 0; i < typeElements.NumField(); i++ {
		tagBodyStr, ok := typeElements.Field(i).Tag.Lookup("redis")
		if !ok {
			return fmt.Errorf("no expected tag name")
		}
		quote := LookUpSingleQuote(tagBodyStr)
		if quote == "" {
			continue
		}
		valueMap[quote] = valueElements.Field(i).Interface()
	}
	// execute
	_, err := redisClient.HMSet(key, valueMap).Result()
	return err
}

func HMGet(redisClient redis.Cmdable, key string, v interface{}) error {
	var hashKeys []string
	typeElements := reflect.TypeOf(v).Elem()
	for i := 0; i < typeElements.NumField(); i++ {
		tagBodyStr, ok := typeElements.Field(i).Tag.Lookup("redis")
		if !ok {
			return fmt.Errorf("no expected tag name")
		}
		quote := LookUpSingleQuote(tagBodyStr)
		if quote == "" {
			continue
		}
		hashKeys = append(hashKeys, quote)
	}
	// 实际查询
	values, err := redisClient.HMGet(key, hashKeys...).Result()
	if err != nil {
		return err
	}
	valueElements := reflect.ValueOf(v).Elem()
	for i := 0; i < valueElements.NumField(); i++ {
		if values[i] == nil {
			return fmt.Errorf("no such field(%s)", hashKeys[i])
		}
		valueStr := values[i].(string)
		elementValue := valueElements.Field(i)
		// 暂时只支持string, int, int32, int64, float32, float64
		switch fieldType := elementValue.Type(); fieldType {
		case reflect.TypeOf(""):
			elementValue.SetString(valueStr)
		case reflect.TypeOf(int(0)), reflect.TypeOf(int32(0)), reflect.TypeOf(int64(0)):
			if valueInt64, err := strconv.ParseInt(valueStr, 10, 64); err != nil {
				return fmt.Errorf("can not transform \"%s\" to int64, error=%+v", valueStr, err)
			} else {
				elementValue.SetInt(valueInt64)
			}
		case reflect.TypeOf(float32(0)), reflect.TypeOf(float64(0)):
			if valueFloat64, err := strconv.ParseFloat(valueStr, 64); err != nil {
				return fmt.Errorf("can not transform \"%s\" to float64, error=%+v", valueStr, err)
			} else {
				elementValue.SetFloat(valueFloat64)
			}
		default:
			return fmt.Errorf("type %+v is not supported", fieldType)
		}
	}
	return nil
}
