package qizhiblock

import (
	"context"
	"github.com/Breeze0806/go-etl/config"
	"github.com/Breeze0806/go-etl/datax/common/plugin"
	"github.com/pingcap/errors"
)

//Job 工作
type Job struct {
	*plugin.BaseJob
	conf *Config
}

//NewJob 创建工作
func NewJob() *Job {
	return &Job{
		BaseJob: plugin.NewBaseJob(),
	}
}

//Init 初始化
func (j *Job) Init(ctx context.Context) (err error) {
	j.conf, err = NewConfig(j.PluginJobConf())
	return errors.Wrapf(err, "NewConfig fail. val: %v", j.PluginJobConf())
}

//Destroy 销毁
func (j *Job) Destroy(ctx context.Context) (err error) {
	return
}

// Split 切分job
func (j *Job) Split(ctx context.Context, number int) ([]*config.JSON, error) {
	return []*config.JSON{j.PluginConf()}, nil
}
