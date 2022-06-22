package main

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

//定义全局变量
type tailTask struct {
	path   string
	topic  string
	tObj   *tail.Tail
	ctx    context.Context
	cancel context.CancelFunc
}

func newTailTask(path string, topic string) *tailTask {
	ctx, cancel := context.WithCancel(context.Background())

	tt := &tailTask{
		path:   path,
		topic:  topic,
		ctx:    ctx,
		cancel: cancel,
	}
	//tt.tObj, err = tail.TailFile(path, cfg)
	return tt
}

func (t *tailTask) Init() (err error) {
	cfg := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	t.tObj, err = tail.TailFile(t.path, cfg)
	return

}

func (t *tailTask) run() (err error) {

	logrus.Infof("collect for path:%s is running...\n", t.path)
	//读取日志发往kafka
	//TailObj ---> log ---clent ---> kafka
	for {
		select {
		case <-t.ctx.Done(): //只要调用t.cannel()就会收到信号
			logrus.Infof("path:%s is stopping...", t.path)
			return
		case line, ok := <-t.tObj.Lines:
			if !ok {
				logrus.Warning("tail file close reopen,path:%s\n", t.path)
				time.Sleep(time.Second) //读出错等一秒
				continue
			}

			//如果是空行就略过.
			if len(strings.Trim(line.Text, "\r")) == 0 {
				logrus.Info("出现空行直接跳过...")
				continue
			}

			//在这里拿到msg数据,利用通道将同步的代码改为异步的.
			//把读出来的一行日志包装成kafka里面的mess的类型,丢到通道中.
			//在单独做一个写入kafka的通道,一个通道tail读,一个通道写kafka,异步进行.
			msg := &sarama.ProducerMessage{}
			msg.Topic = t.topic
			msg.Value = sarama.StringEncoder(line.Text)
			//丢到管道中
			ToMsgChan(msg)
			//msgChan <- msg //<---管道暴露了..

		}

	}
}
