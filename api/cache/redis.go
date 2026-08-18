package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

const cacheExpiry time.Duration = 24 * time.Hour // TODO: properly set this!
const redisClientContextKey = "redisClientContextKey"

func getClient(ctx context.Context) *redis.Client {
	cli, ok := ctx.Value(redisClientContextKey).(*redis.Client)
	if !ok {
		return nil
	}
	return cli
}

// TODO: consider using!
func NewRedisClient(ctx context.Context, hostAndPort string, password string) (context.Context, *redis.Client, error) {
	// TODO: connecting with opts
	opts := &redis.Options{
		Addr:     hostAndPort, // Server address
		Password: password,    // No password by default ""
		DB:       0,           // Default database (0-15)
		Protocol: 3,           // RESP3 protocol (recommended for v9) // TODO: ???
	}
	//// TODO: TLS conn
	//opts := &redis.Options{
	//	Addr:     hostAndPort,
	//	Password: password,
	//	TLSConfig: &tls.Config{
	//		MinVersion: tls.VersionTLS12, // Enforce TLS 1.2+
	//	},
	//}
	//// TODO: connecting with URL
	//// Format: redis://<user>:<pass>@<host>:<port>/<db>
	//opts, err := redis.ParseURL("redis://user:password@" + hostAndPort + "/0")
	//if err != nil {
	//	panic(err)
	//}
	//// TODO: connecting to a cluster
	////rdb := redis.NewClusterClient(&redis.ClusterOptions{
	////	Addrs: []string{
	////		"node1.cluster.local:6379",
	////		"node2.cluster.local:6379",
	////		"node3.cluster.local:6379",
	////	},
	////	Password: "cluster-password",
	////})
	//
	////return redis.NewClient(opt)
	//redis.NewClient(&redis.Options{
	//	Addr:     "localhost:6379", // Redis address
	//	Password: "",               // No password by default
	//	DB:       0,                // Default database
	//})
	//redis.Option
	client := redis.NewClient(opts)
	return context.WithValue(ctx, redisClientContextKey, client), client, nil
	//client := redis.NewClient(&redis.Options{
	//	Network:                    hostAndPort,
	//	Addr:                       hostAndPort,
	//	ClientName:                 "",
	//	Dialer:                     nil,
	//	OnConnect:                  nil,
	//	Protocol:                   0,
	//	Username:                   "",
	//	Password:                   "",
	//	CredentialsProvider:        nil,
	//	CredentialsProviderContext: nil,
	//	DB:                         0,
	//	MaxRetries:                 0,
	//	MinRetryBackoff:            0,
	//	MaxRetryBackoff:            0,
	//	DialTimeout:                0,
	//	ReadTimeout:                0,
	//	WriteTimeout:               0,
	//	ContextTimeoutEnabled:      false,
	//	PoolFIFO:                   false,
	//	PoolSize:                   0,
	//	PoolTimeout:                0,
	//	MinIdleConns:               0,
	//	MaxIdleConns:               0,
	//	MaxActiveConns:             0,
	//	ConnMaxIdleTime:            0,
	//	ConnMaxLifetime:            0,
	//	TLSConfig:                  nil,
	//	Limiter:                    nil,
	//	DisableIndentity:           false,
	//	DisableIdentity:            false,
	//	IdentitySuffix:             "",
	//	UnstableResp3:              false,
	//})
}

func Get(ctx context.Context, key string) ([]byte, error) {
	client := getClient(ctx)
	return client.Get(ctx, key).Bytes()
}

func Set(ctx context.Context, key string, val []byte) error {
	client := getClient(ctx)
	return client.Set(ctx, key, val, cacheExpiry).Err()
}

func Del(ctx context.Context, key string) error {
	client := getClient(ctx)
	return client.Del(ctx, key).Err()
}
