package qizhiblock

import (
	"context"
	"github.com/Breeze0806/go-etl/datax/common/plugin"
	"github.com/Breeze0806/go-etl/datax/common/spi/writer"
	"github.com/Breeze0806/go-etl/datax/core/transport/exchange"
	"github.com/Breeze0806/go-etl/element"
	"github.com/pingcap/errors"
	"sync"
)

//Task 任务
type Task struct {
	*writer.BaseTask
	conf         *Config
	restApiToken string
	client       *apiClient
}

func (t *Task) Init(ctx context.Context) (err error) {
	if t.conf, err = NewConfig(t.PluginJobConf()); err != nil {
		return t.Wrapf(err, "newConfig fail")
	}

	if t.conf.Username == "" || t.conf.Password == "" || t.conf.RestApiEndpoint == "" || t.conf.Channel == "" {
		return errors.New("username,password,restApiEndpoint,channel can't be empty")
	}
	if t.conf.ChainCode == "" {
		t.conf.ChainCode = "map"
	}
	if t.conf.ChainCodeFunction == "" {
		t.conf.ChainCodeFunction = "put"
	}
	if t.client, err = NewApiClient(t.conf); err != nil {
		log.Errorf("failed to auth to %s by credential %s %s", t.conf.RestApiEndpoint, t.conf.Username, t.conf.Password)
		return t.Wrapf(err, "newConfig fail")
	}
	return
}

func (t *Task) Destroy(ctx context.Context) error {
	return nil
}

func (t *Task) StartWrite(ctx context.Context, receiver plugin.RecordReceiver) error {
	recordChan := make(chan element.Record)
	var wg sync.WaitGroup
	var rerr error
	wg.Add(1)
	afterCtx, cancel := context.WithCancel(ctx)
	//通过该协程读取记录接受器receiver的记录放入recordChan
	go func() {
		defer func() {
			wg.Done()
			//关闭recordChan
			close(recordChan)
			log.Debugf(t.Format("get records end"))
		}()
		log.Debugf(t.Format("start to get records"))
		for {
			select {
			case <-afterCtx.Done():
				return
			default:
			}
			var record element.Record
			record, rerr = receiver.GetFromReader()
			if rerr != nil && rerr != exchange.ErrEmpty {
				return
			}

			//当记录接受器receiver返回不为空错误，写入recordChan
			if rerr != exchange.ErrEmpty {
				select {
				//防止在ctx关闭时不写入recordChan
				case <-afterCtx.Done():
					return
				case recordChan <- record:
				}

			}
		}
	}()
	log.Debugf(t.Format("start to write"))
	var err error
	for {
		select {
		case record, ok := <-recordChan:
			if !ok {
				log.Infof("write is over")
				if err == nil {
					err = rerr
				}
				goto End
			}
			if err = t.client.writeRecord(t.conf.KeyColumn, record); err != nil {
				log.Errorf("failed to write record %v", err)
				goto End
			}
		}
	}
End:
	// 写入完毕后，关闭结果读取
	cancel()
	log.Debugf(t.Format("wait all goroutine"))
	//等待携程结束
	wg.Wait()
	log.Debugf(t.Format(" wait all goroutine end"))
	switch {
	//当外部取消时，开始写入不是错误
	case ctx.Err() != nil:
		return nil
	//当错误是停止时，也不是错误
	case err == exchange.ErrTerminate:
		return nil
	}
	return t.Wrapf(err, "")
}

func NewTask() *Task {
	return &Task{
		BaseTask: writer.NewBaseTask(),
	}
}
