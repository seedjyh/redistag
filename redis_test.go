// redis_test
package redistag

import (
	"github.com/go-redis/redis"
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
	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:              testAddressList,
	})
	assert.NoError(t, HMSet(redisClient, key, raw))
	readResult := &NumberPO{}
	assert.NoError(t, HMGet(redisClient, key, readResult))
	assert.Equal(t, raw, readResult)
}
