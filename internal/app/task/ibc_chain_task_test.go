package task

import (
	"context"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/conf"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/repository"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/repository/cache"
	"testing"
)

func TestMain(m *testing.M) {
	cache.InitRedisClient(conf.Redis{
		Addrs:    "127.0.0.1:6379",
		User:     "",
		Password: "",
		Mode:     "single",
		Db:       0,
	})
	repository.InitMgo(conf.Mongo{
		Url:      "mongodb://ibc:ibcpassword@192.168.150.60:27018/?connect=direct&authSource=iobscan-ibc",
		Database: "iobscan-ibc",
	}, context.Background())
	m.Run()
}

func TestRunOnce(t *testing.T) {
	new(IbcChainCronTask).Run()
}