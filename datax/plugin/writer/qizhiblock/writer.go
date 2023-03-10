package qizhiblock

import (
	"github.com/Breeze0806/go-etl/config"
	spiwriter "github.com/Breeze0806/go-etl/datax/common/spi/writer"
)

// Writer 数据上链写入
type Writer struct {
	pluginConf *config.JSON
}

//ResourcesConfig 插件资源配置
func (w *Writer) ResourcesConfig() *config.JSON {
	return w.pluginConf
}

//Job 工作
func (w *Writer) Job() spiwriter.Job {
	job := NewJob()
	job.SetPluginConf(w.pluginConf)
	return job
}

//Task 任务
func (w *Writer) Task() spiwriter.Task {
	task := NewTask()
	task.SetPluginConf(w.pluginConf)
	return task
}
