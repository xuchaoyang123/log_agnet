package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
)

type Config struct {
	KafkaConfig   `ini:"kafka"`
	CollectConfig `ini:"collect"`
	EtcdConfig    `ini:"etcd"`
}
type KafkaConfig struct {
	Address  string `ini:"address"`
	Topic    string `int:"topic"`
	ChanSize int64  `int:"chan_size"`
}

type CollectConfig struct {
	LogFilePath string `ini:"logfile_path"`
}

type EtcdConfig struct {
	Address    string `ini:"address"`
	CollectKey string `ini:"collect_key"`
}

func run() {
	select {}
}

func main3() {

	//先获取ip在读配置文件 为后续去etcd配置打下基础
	ip, err2 := GetOutBoundIp()
	if err2 != nil {
		logrus.Errorf("get ip failed, err:%v", err2)
		return
	}

	var configObj = new(Config)
	//1.读取配置文件go-ini
	err := ini.MapTo(configObj, "./conf.ini")
	if err != nil {
		logrus.Errorf("load config failed,err:%v", err)
		return
	}
	fmt.Println(configObj)

	//1. 初始化链接kafka
	err = InitKafka([]string{configObj.KafkaConfig.Address}, configObj.KafkaConfig.ChanSize)
	if err != nil {
		logrus.Errorf("init kafka failed, err %v", err)
		return
	}
	logrus.Info("init kafka success!")

	//初始化etcd连接
	err = InitEtcd([]string{configObj.EtcdConfig.Address})
	if err != nil {
		logrus.Errorf("init etcd failed, err %v", err)
		return
	}
	//从etcd中拉取要收集的日志配置项
	collectKey := fmt.Sprintf(configObj.EtcdConfig.CollectKey, ip)
	allConf, err := EtcdGetConf(collectKey)
	if err != nil {
		logrus.Errorf("get conf from etcd failed, err %v", err)
		return
	}
	fmt.Println(allConf)
	//派一个小弟去监控etcd中configObj.EtcdConfig.CollectKey
	go EtcdWatchConf(collectKey)

	//2.根据配置中的日志路径初始化tail
	err = InitTailFile(allConf) //etcd中获取的配置项传到init中
	if err != nil {
		logrus.Errorf("init tailFile failed, err %v", err)
		return
	}
	logrus.Info("init tailFile success!")
	run()

}
