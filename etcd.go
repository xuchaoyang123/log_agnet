package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

//etcd 相关操作

var (
	Client *clientv3.Client
)

func InitEtcd(address []string) (err error) {

	Client, err = clientv3.New(clientv3.Config{
		Endpoints:         address,
		DialKeepAliveTime: time.Second * 5,
	})

	if err != nil {
		fmt.Printf("connect to etcd failed err:%v\n", err)
		return
	}
	return
}

//拉取日志收集配置项的函数
func EtcdGetConf(key string) (colletcdEntryList []ColletcdEntry, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	resp, err := Client.Get(ctx, key)
	if err != nil {
		logrus.Errorf("get conf from etcd by key :%s failed err:%s", key, err)
		return
	}

	if len(resp.Kvs) == 0 {
		logrus.Warningf("get len:0 from etcd by key :%s ", key)
		return
	}

	//json格式字符串
	err = json.Unmarshal(resp.Kvs[0].Value, &colletcdEntryList)
	if err != nil {
		logrus.Errorf("json Unmarshal failed,err :%s", err)
		return
	}
	return
}

//监控etcd中日志收集配置变化的函数
func EtcdWatchConf(key string) {

	//循环监视
	for {
		watchCh := Client.Watch(context.Background(), key)

		for wresp := range watchCh {
			logrus.Info("get new conf from etcd!")
			for _, evt := range wresp.Events {
				fmt.Printf("type:%s,key:%s,values:%s\n", evt.Type, evt.Kv.Key, evt.Kv.Value)
				var newConf []ColletcdEntry
				if evt.Type == clientv3.EventTypeDelete {
					//如果是删除事件
					logrus.Warning("😞 FBI etcd delete the key!!")
					SendNewConf(newConf)
					continue
				}
				err := json.Unmarshal(evt.Kv.Value, &newConf)
				if err != nil {
					logrus.Errorf("😭 json Unmarshal new conf failed,err %s!", err)
					continue
				}

				//告诉tailFile模块应该启用新的配置
				SendNewConf(newConf) //执行到这步就堵塞了
			}
		}

	}

}
