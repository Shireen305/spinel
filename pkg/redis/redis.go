package redis

import (
	"context"
	"fmt"
	"os"
	"sync"

	. "github.com/souvikdeyrit/spinel/pkg/meta"
	"github.com/souvikdeyrit/spinel/pkg/utils"

	"github.com/go-redis/redis/v8"
)

var logger = utils.GetLogger("spinel")

type RedisConfig struct {
	Strict  bool // update ctime
	Retries int
}

// struct to store Redis specific data of metadata engine (implements Meta interface)
type redisMeta struct {
	sync.Mutex
	conf *RedisConfig
	rdb  *redis.Client

	sid          int64
	openFiles    map[Ino]int
	removedFiles map[Ino]bool
	msgCallbacks *msgCallbacks
}

type msgCallbacks struct {
	sync.Mutex
	callbacks map[uint32]MsgCallback
}

var c = context.TODO()

func NewRedisMeta(url string, conf *RedisConfig) (Meta, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %s", url, err)
	}
	if opt.Password == "" && os.Getenv("REDIS_PASSWD") != "" {
		opt.Password = os.Getenv("REDIS_PASSWD")
	}
	m := &redisMeta{
		conf:         conf,
		rdb:          redis.NewClient(opt),
		openFiles:    make(map[Ino]int),
		removedFiles: make(map[Ino]bool),
		msgCallbacks: &msgCallbacks{
			callbacks: make(map[uint32]MsgCallback),
		},
	}
	m.sid, err = m.rdb.Incr(c, "nextsession").Result()
	if err != nil {
		return nil, fmt.Errorf("create session: %s", err)
	}
	logger.Debugf("session is in %d", m.sid)
	go m.refreshSession()
	go m.cleanupChunks() // Metadata engine garbage collector process
	return m, nil // Ignore the error, RedisMeta needs to implement the Meta interface, later
}

// Later
func (r *redisMeta) refreshSession() {
	
}

// Later
func (r *redisMeta) cleanupChunks() {
}
