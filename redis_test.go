// redis_test
package redistag

import (
	"context"
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
)

// 如果没有可用的本地redis集群，将 TestWithRealRedisClusterFlag 设为false，可以跳过需要真实redis集群的测试用例。
var TestWithRealRedisClusterFlag = true

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

func prepareClientForTest(t *testing.T) *redisV8.ClusterClient {
	if !TestWithRealRedisClusterFlag {
		t.SkipNow()
	}
	return redisV8.NewClusterClient(&redisV8.ClusterOptions{
		Addrs: testAddressList,
	})
}

// 相同格式，先写入再读出
func TestRedisClient(t *testing.T) {
	redisClient := prepareClientForTest(t)
	ctx := context.Background()
	redisClient.Del(ctx, key)
	defer redisClient.Del(ctx, key)
	assert.NoError(t, HMSet(ctx, redisClient, key, moreNumber))
	readResult := &MoreNumberPO{}
	assert.NoError(t, HMGet(ctx, redisClient, key, readResult))
	assert.Equal(t, moreNumber, readResult)
}

// 按字段少的结构新增，按字段多的结构读出。
func TestHMGet(t *testing.T) {
	redisClient := prepareClientForTest(t)
	ctx := context.Background()
	redisClient.Del(ctx, key)
	defer redisClient.Del(ctx, key)
	assert.NoError(t, HMSet(ctx, redisClient, key, lessNumber))
	readResult := &MoreNumberPO{}
	assert.NoError(t, HMGet(ctx, redisClient, key, readResult))
	assert.Equal(t, lessNumber.Id, readResult.Id)
	assert.Equal(t, lessNumber.Msisdn, readResult.Msisdn)
	assert.Equal(t, "", readResult.AreaCode)
	assert.Equal(t, 0, readResult.LocationState)
	assert.EqualValues(t, 0.0, readResult.Balance)
}

// 按字段多的结构新增，按字段少的结构读出。
func TestHMGet2(t *testing.T) {
	redisClient := prepareClientForTest(t)
	ctx := context.Background()
	redisClient.Del(ctx, key)
	defer redisClient.Del(ctx, key)
	assert.NoError(t, HMSet(ctx, redisClient, key, moreNumber))
	readResult := &LessNumberPO{}
	assert.NoError(t, HMGet(ctx, redisClient, key, readResult))
	assert.Equal(t, moreNumber.Id, readResult.Id)
	assert.Equal(t, moreNumber.Msisdn, readResult.Msisdn)
}

// 按字段多的新增，按字段少的再度写入，按字段少的读取。
func TestHMSet(t *testing.T) {
	redisClient := prepareClientForTest(t)
	ctx := context.Background()
	redisClient.Del(ctx, key)
	defer redisClient.Del(ctx, key)
	assert.NoError(t, HMSet(ctx, redisClient, key, moreNumber))
	assert.NoError(t, HMSet(ctx, redisClient, key, lessNumber))
	readResult := &LessNumberPO{}
	assert.NoError(t, HMGet(ctx, redisClient, key, readResult))
	assert.Equal(t, lessNumber, readResult)
}

// 按字段少的新增，按字段多的再度写入，按字段多的读取。
func TestHMSet2(t *testing.T) {
	redisClient := prepareClientForTest(t)
	ctx := context.Background()
	redisClient.Del(ctx, key)
	defer redisClient.Del(ctx, key)
	assert.NoError(t, HMSet(ctx, redisClient, key, lessNumber))
	assert.NoError(t, HMSet(ctx, redisClient, key, moreNumber))
	readResult := new(MoreNumberPO)
	assert.NoError(t, HMGet(ctx, redisClient, key, readResult))
	assert.Equal(t, moreNumber, readResult)
}

// 结构体某些字段没有带redis标签。
func TestHMSet3(t *testing.T) {
	redisClient := prepareClientForTest(t)
	ctx := context.Background()
	redisClient.Del(ctx, key)
	defer redisClient.Del(ctx, key)
	type Foo struct {
		Title     string  `redis:"'title123'"`
		Price     float32 `redis:"'price456'"`
		NoTag     string
		LastField string `redis:"'last-field'"`
	}
	foo := &Foo{
		Title:     "this-is-title",
		Price:     3.14,
		NoTag:     "this-is-no-tag",
		LastField: "this-is-last-field",
	}
	assert.NoError(t, HMSet(ctx, redisClient, key, foo))
	readResult := new(Foo)
	assert.NoError(t, HMGet(ctx, redisClient, key, readResult))
	assert.Equal(t, foo.Title, readResult.Title)
	assert.Equal(t, foo.Price, readResult.Price)
	assert.Equal(t, "", readResult.NoTag)
	assert.Equal(t, foo.LastField, readResult.LastField)
}
