package worker

import (
	db "github.com/elkelk/simplebank/db/sqlc"
	"github.com/hibiken/asynq"
)

type TaskProcessor interface {
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

func NewRedisProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{},
	)

	return &RedisTaskProcessor{
		server: server,
		store:  store,
	}
}
