// redis_test
package redistag

import (
	redisV7 "github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testAddressList = []string{
	"localhost:8001",
	"localhost:8002",
	"localhost:8003",
	"localhost:8004",
	"localhost:8005",
	"localhost:8006",
}

type NumberPO struct {
	Id int64 `redis:"'id'"`
	Msisdn string `redis:"'msisdn'"`
	AreaCode string `redis:"'areacode'"`
	LocationState int `redis:"'locationstate'"`
	Balance float32 `redis:"'balance'"`
}

func TestLookUpSingleQuote(t *testing.T) {
	assert.Equal(t, "name", LookUpSingleQuote("'name'"))
	assert.Equal(t, "name", LookUpSingleQuote("a'name'b'c'd"))
	assert.Equal(t, "", LookUpSingleQuote("''"))
	assert.Equal(t, "", LookUpSingleQuote("a'b"))
}

func TestRedisClient(t *testing.T) {
	const BASE_KEY = "test-tag:"
	raw := &NumberPO{
		Id:            10000000000000,
		Msisdn:        "13700112233",
		AreaCode:      "0571",
		LocationState: 2,
		Balance: 3.1415926,
	}
	key := BASE_KEY + raw.Msisdn
	redisClient := redisV7.NewClusterClient(&redisV7.ClusterOptions{
		Addrs:              testAddressList,
	})
	assert.NoError(t, HMSet(redisClient, key, raw))
	readResult := &NumberPO{}
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, raw, readResult)
}

// 数据库里的hset的字段比想要读取的少。
func TestHMGet(t *testing.T) {
	const BASE_KEY = "test-tag:"
	type LessNumberPO struct {
		Id int64 `redis:"'id'"`
		Msisdn string `redis:"'msisdn'"`
	}
	redisClient := redisV7.NewClusterClient(&redisV7.ClusterOptions{
		Addrs:              testAddressList,
	})
	key := BASE_KEY + "write-less-read-more"
	// 删除键
	redisClient.Del(key)
	// 把字段少的写入。
	raw := &LessNumberPO{
		Id:            10000000000000,
		Msisdn:        "13700112233",
	}
	assert.NoError(t, HMSet(redisClient, key, raw))
	// 按字段多的读取
	readResult := &NumberPO{}
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, raw.Id, readResult.Id)
	assert.Equal(t, raw.Msisdn, readResult.Msisdn)
	assert.Equal(t, "", readResult.AreaCode)
	assert.Equal(t, 0, readResult.LocationState)
	assert.EqualValues(t, 0.0, readResult.Balance)
}

// 数据库里的hset的字段比想要读取的多。
func TestHMGet2(t *testing.T) {
	const BASE_KEY = "test-tag:"
	type MoreNumberPO struct {
		Id int64 `redis:"'id'"`
		Msisdn string `redis:"'msisdn'"`
		AreaCode string `redis:"'areacode'"`
		LocationState int `redis:"'locationstate'"`
		Balance float32 `redis:"'balance'"`
		UserData string `redis:"'userData'"`
	}
	redisClient := redisV7.NewClusterClient(&redisV7.ClusterOptions{
		Addrs:              testAddressList,
	})
	key := BASE_KEY + "write-more-read-less"
	// 删除键
	redisClient.Del(key)
	// 把字段多的写入。
	raw := &MoreNumberPO{
		Id:            10000000000000,
		Msisdn:        "13700112233",
		AreaCode:      "0571",
		LocationState: 2,
		Balance:       3.1415926,
		UserData:      "some-data",
	}
	assert.NoError(t, HMSet(redisClient, key, raw))
	// 按字段少的读取
	readResult := &NumberPO{}
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, raw.Id, readResult.Id)
	assert.Equal(t, raw.Msisdn, readResult.Msisdn)
	assert.Equal(t, raw.AreaCode, readResult.AreaCode)
	assert.Equal(t, raw.LocationState, readResult.LocationState)
	assert.EqualValues(t, raw.Balance, readResult.Balance)
}

// 原来就有hset，但新的字段更少
func TestHMSet(t *testing.T) {
	const BASE_KEY = "test-tag:"
	type LessNumberPO struct {
		Id int64 `redis:"'id'"`
		Msisdn string `redis:"'msisdn'"`
	}
	redisClient := redisV7.NewClusterClient(&redisV7.ClusterOptions{
		Addrs:              testAddressList,
	})
	key := BASE_KEY + "write-more-write-less"
	// 删除键
	redisClient.Del(key)
	// 把字段多的写入。
	raw := &NumberPO{
		Id:            10000000000000,
		Msisdn:        "13700112233",
		AreaCode:      "",
		LocationState: 0,
		Balance:       0,
	}
	assert.NoError(t, HMSet(redisClient, key, raw))
	// 按字段少的再写入
	write2 := &LessNumberPO{
		Id:     123,
		Msisdn: "some",
	}
	assert.NoError(t, HMSet(redisClient, key, write2))
	// 确认写入效果
	readResult := &LessNumberPO{}
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, write2, readResult)
}

// 原来就有hset，但新的字段更多
func TestHMSet2(t *testing.T) {
	const BASE_KEY = "test-tag:"
	type MoreNumberPO struct {
		Id int64 `redis:"'id'"`
		Msisdn string `redis:"'msisdn'"`
		AreaCode string `redis:"'areacode'"`
		LocationState int `redis:"'locationstate'"`
		Balance float32 `redis:"'balance'"`
		UserData string `redis:"'userData'"`
	}
	redisClient := redisV7.NewClusterClient(&redisV7.ClusterOptions{
		Addrs:              testAddressList,
	})
	key := BASE_KEY + "write-more-write-less"
	// 删除键
	redisClient.Del(key)
	// 把字段少的写入。
	raw := &NumberPO{
		Id:            10000000000000,
		Msisdn:        "13700112233",
		AreaCode:      "",
		LocationState: 0,
		Balance:       0,
	}
	assert.NoError(t, HMSet(redisClient, key, raw))
	// 按字段多的再写入
	write2 := &MoreNumberPO{
		Id:            1212,
		Msisdn:        "some",
		AreaCode:      "test",
		LocationState: 123,
		Balance:       56.7,
		UserData:      "ok",
	}
	assert.NoError(t, HMSet(redisClient, key, write2))
	// 确认写入效果
	readResult := new(MoreNumberPO)
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, write2, readResult)
}
