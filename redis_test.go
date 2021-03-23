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

type LessNumberPO struct {
	Id     int64  `redis:"'id'"`
	Msisdn string `redis:"'msisdn'"`
}

type MoreNumberPO struct {
	Id            int64   `redis:"'id'"`
	Msisdn        string  `redis:"'msisdn'"`
	AreaCode      string  `redis:"'areacode'"`
	LocationState int     `redis:"'locationstate'"`
	Balance       float32 `redis:"'balance'"`
	UserData      string  `redis:"'userData'"`
}

var lessNumber = &LessNumberPO{
	Id:     123,
	Msisdn: "13700112233",
}
var moreNumber = &MoreNumberPO{
	Id:            789,
	Msisdn:        "19966778899",
	AreaCode:      "6789",
	LocationState: 9,
	Balance:       999.99,
	UserData:      "{\"there\":\"are\", \"more\":\"fields\"}",
}
const key = "redistag:test-tag:abc001"

func TestLookUpSingleQuote(t *testing.T) {
	assert.Equal(t, "name", LookUpSingleQuote("'name'"))
	assert.Equal(t, "name", LookUpSingleQuote("a'name'b'c'd"))
	assert.Equal(t, "", LookUpSingleQuote("''"))
	assert.Equal(t, "", LookUpSingleQuote("a'b"))
}

// 相同格式，先写入再读出
func TestRedisClient(t *testing.T) {
	redisClient := redisV7.NewClusterClient(&redisV7.ClusterOptions{
		Addrs: testAddressList,
	})
	redisClient.Del(key)
	defer redisClient.Del(key)
	assert.NoError(t, HMSet(redisClient, key, moreNumber))
	readResult := &MoreNumberPO{}
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, moreNumber, readResult)
}

// 按字段少的结构新增，按字段多的结构读出。
func TestHMGet(t *testing.T) {
	redisClient := redisV7.NewClusterClient(&redisV7.ClusterOptions{
		Addrs: testAddressList,
	})
	redisClient.Del(key)
	defer redisClient.Del(key)
	assert.NoError(t, HMSet(redisClient, key, lessNumber))
	readResult := &MoreNumberPO{}
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, lessNumber.Id, readResult.Id)
	assert.Equal(t, lessNumber.Msisdn, readResult.Msisdn)
	assert.Equal(t, "", readResult.AreaCode)
	assert.Equal(t, 0, readResult.LocationState)
	assert.EqualValues(t, 0.0, readResult.Balance)
}

// 按字段多的结构新增，按字段少的结构读出。
func TestHMGet2(t *testing.T) {
	redisClient := redisV7.NewClusterClient(&redisV7.ClusterOptions{
		Addrs: testAddressList,
	})
	redisClient.Del(key)
	defer redisClient.Del(key)
	assert.NoError(t, HMSet(redisClient, key, moreNumber))
	readResult := &LessNumberPO{}
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, moreNumber.Id, readResult.Id)
	assert.Equal(t, moreNumber.Msisdn, readResult.Msisdn)
}

// 按字段多的新增，按字段少的再度写入，按字段少的读取。
func TestHMSet(t *testing.T) {
	redisClient := redisV7.NewClusterClient(&redisV7.ClusterOptions{
		Addrs: testAddressList,
	})
	redisClient.Del(key)
	defer redisClient.Del(key)
	assert.NoError(t, HMSet(redisClient, key, moreNumber))
	assert.NoError(t, HMSet(redisClient, key, lessNumber))
	readResult := &LessNumberPO{}
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, lessNumber, readResult)
}

// 按字段少的新增，按字段多的再度写入，按字段多的读取。
func TestHMSet2(t *testing.T) {
	redisClient := redisV7.NewClusterClient(&redisV7.ClusterOptions{
		Addrs: testAddressList,
	})
	redisClient.Del(key)
	defer redisClient.Del(key)
	assert.NoError(t, HMSet(redisClient, key, lessNumber))
	assert.NoError(t, HMSet(redisClient, key, moreNumber))
	readResult := new(MoreNumberPO)
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, moreNumber, readResult)
}
