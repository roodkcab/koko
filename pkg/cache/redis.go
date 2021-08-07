package cache

import (
	"github.com/jumpserver/koko/pkg/config"
	"github.com/mediocregopher/radix/v3"
	"net"
	"time"
)

type Cache struct {
	pool *radix.Pool
}

type Config struct {
	// Defaults to "tcp".
	Network string
	// Addr of a single redis server instance.
	// See "Clusters" field for clusters support.
	// Defaults to "127.0.0.1:6379".
	Addr string
	// Clusters a list of network addresses for clusters.
	// If not empty "Addr" is ignored.
	Clusters []string

	Password    string
	DialTimeout time.Duration

	// MaxActive defines the size connection pool.
	// Defaults to 10.
	MaxActive int

	DBIndex int
}

func Init() (c *Cache, err error) {
	conf := config.GetConf()
	var dialOptions []radix.DialOpt
	dialOptions = append(dialOptions, radix.DialAuthPass(conf.RedisPassword))
	//dialOptions = append(dialOptions, radix.DialTimeout(cfg.DialTimeout))
	dialOptions = append(dialOptions, radix.DialSelectDB(int(conf.RedisDBIndex)))
	var connFunc radix.ConnFunc
	connFunc = func(network, addr string) (radix.Conn, error) {
		return radix.Dial("tcp", net.JoinHostPort(conf.RedisHost, conf.RedisPort), dialOptions...)
	}
	pool, err := radix.NewPool("", "", 30, radix.PoolConnFunc(connFunc))
	return &Cache{pool: pool}, err
}

func (c *Cache) Get(key string) (res []byte, err error) {
	err = c.pool.Do(radix.Cmd(&res, "GET", key))
	return res, err
}

func (c *Cache) Set(key string, val []byte, expire int) (err error) {
	err = c.pool.Do(radix.FlatCmd(nil, "SET", key, val))
	if err != nil {
		return
	}
	err = c.pool.Do(radix.FlatCmd(nil, "EXPIRE", key, expire))
	return
}