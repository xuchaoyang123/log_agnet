package main

import (
	"github.com/sirupsen/logrus"
)

//tailTask管理者
type tailTaskMgr struct {
	tailTaskMap      map[string]*tailTask //所有的tailtask任务
	collectEntryList []ColletcdEntry      //所有配置项
	confChan         chan []ColletcdEntry //等待新配置的通道
}

var (
	ttMgr *tailTaskMgr
)

func InitTailFile(allConf []ColletcdEntry) (err error) {
	//allconf 里面存了好多个path路径文件.
	//针对每一个日志收集项创建一个对应的tailobj
	ttMgr = &tailTaskMgr{
		tailTaskMap:      make(map[string]*tailTask, 20),
		collectEntryList: allConf,
		confChan:         make(chan []ColletcdEntry),
	}

	for _, conf := range allConf {
		tt := newTailTask(conf.Path, conf.Topic) //创建一个日志收集任务
		err = tt.Init()                          //创建一个日志收集任务
		if err != nil {
			logrus.Error("create tailobj for  path: %s failed,err:%v\n", conf.Path, err)
			continue
		}
		logrus.Infof("create a tail task  for path:%s auccess\n", conf.Path)
		ttMgr.tailTaskMap[tt.path] = tt //把创建的这个tailTask任务登记一下方便后续管理.
		//启一个后台的go去收集日志吧
		go tt.run()

	}

	go ttMgr.Watch() //在后台等新的配置来
	return

}

func (t *tailTaskMgr) Watch() {

	for {
		//派一个小弟等着新配置来
		newConf := <-t.confChan //取到数值说明新的配置来了.
		//新配置来了之后,应该管理一下我之前的启动的tailTask
		logrus.Info("get new conf from etcd,conf:%s,start manage tailTask..\n", newConf)
		for _, conf := range newConf {
			//1. 原来已经存在的任务就不用动了。
			if t.isExist(conf) {
				continue
			}
			//2. 原来没有他新创建一个tailTask任务
			tt := newTailTask(conf.Path, conf.Topic) //创建一个日志收集任务
			err := tt.Init()                         //创建一个日志收集任务
			if err != nil {
				logrus.Error("create tailobj for  path: %s failed,err:%v\n", conf.Path, err)
				continue
			}
			logrus.Infof("create a tail task  for path:%s auccess\n", conf.Path)
			//启一个后台的go去收集日志吧
			go tt.run()
		}
		//3.原来有现在没有的要停掉.
		//找出tailTaskMap中存在但是newConf不存在的哪些tailTask,把他们关掉.
		for key, task := range t.tailTaskMap {
			var found bool
			for _, conf := range newConf {
				if key == conf.Path {
					found = true
					break
				}
			}
			if !found {
				//拿当前的task任务信息和之前存的task任务json数据做对比
				//找不到的哪项path就要停掉他.
				logrus.Infof("the task collect path:%s need to stop.", task.path)
				delete(t.tailTaskMap, key) //从管理类中删除信息
				task.cancel()

			}
		}
	}
}

//判断
func (t *tailTaskMgr) isExist(conf ColletcdEntry) bool {
	_, ok := t.tailTaskMap[conf.Path]
	return ok
}

func SendNewConf(newConf []ColletcdEntry) {
	ttMgr.confChan <- newConf

}
