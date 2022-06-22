package main

import (
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

var (
	client  sarama.SyncProducer
	msgChan chan *sarama.ProducerMessage //定义要写入kafka的通道
)

//初始化全局的kafka clent
func InitKafka(address []string, chanSize int64) (err error) {

	//1. 生产者配置
	config := sarama.NewConfig()
	//配置kafka主从响应是必须all模式,必须都同步了才会进行ack响应
	config.Producer.RequiredAcks = sarama.WaitForAll
	//分区
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	//确认
	config.Producer.Return.Successes = true

	//2.链接kafka 全局的client
	client, err = sarama.NewSyncProducer(address, config)
	if err != nil {
		logrus.Error("producer closed, err:", err)
		return
	}
	//初始化msgChan
	msgChan = make(chan *sarama.ProducerMessage, chanSize)
	//起一个后台的goroutine从msgchan中读数据
	go sendMsg()
	return
}

//从msgChan中读取msg,发送给kafka
func sendMsg() {
	for {
		select {
		case msg := <-msgChan:
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				logrus.Warning("send msg failed,err:", err)
				return
			}
			logrus.Infof("send msg to kafka succeess. pid:%v offset:%v", pid, offset)
		}
	}
}

////定义一个函数,向外暴露msgChan
func ToMsgChan(msg *sarama.ProducerMessage) {
	msgChan <- msg
}
