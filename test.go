package main

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"log"
	"net"
	"time"
)

var (
	client1  influxdb2.Client
	writeAPI api.WriteAPI
	datas    *write.Point
)

func GetIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localaddr := conn.LocalAddr().(*net.UDPAddr)
	//fmt.Println(localaddr.String()) //ip:port
	ip := localaddr.IP.String() //ip
	return ip
}

//cpu Percent
func GetCpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return percent[0]
}

//mem
type Men struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

func GetMem() *Men {
	var memInfo = new(Men)
	info, _ := mem.VirtualMemory()
	memInfo.Total = info.Total
	memInfo.Free = info.Free
	memInfo.Available = info.Available
	memInfo.Used = info.Used
	memInfo.UsedPercent = info.UsedPercent

	return memInfo
}

//disk

type DiskInfo struct {
	p map[string]*disk.UsageStat
}

func GetDisk() *DiskInfo {
	var diskinfo = &DiskInfo{p: make(map[string]*disk.UsageStat, 16)}

	//获取所有分区信息  parts//是个list列表
	parts, _ := disk.Partitions(true)
	for _, part := range parts { /// // 				<--------为啥parts 不能进行循环遍历,好像是和在func里有关系.
		fmt.Println("---", part)
		//拿到每个磁盘分区信息
		disk, _ := disk.Usage(part.Mountpoint) //传入挂载点进去
		//写入数据
		diskinfo.p[part.Mountpoint] = disk
		return diskinfo

	}
	////磁盘io
	//ioStat, _ := disk.IOCounters()
	//for k, v := range ioStat {
	//	fmt.Println(k, v)
	//}
	return nil
}

func ConnInfluxdb2(token string, url string) {
	//配置连接
	client1 = influxdb2.NewClient(url, token)
	defer client1.Close()
}

func WriteInfluxdb2(org string, bucket string, datas *write.Point) {

	writeAPI = client1.WriteAPI(org, bucket)
	// write point asynchronously
	writeAPI.WritePoint(datas)
	// Flush writes
	writeAPI.Flush()
}

func ReadInfluxdb2(org string, bucket string, measurement string) {

	query := fmt.Sprintf("from(bucket:\"%v\")|> range(start: -1h) |> filter(fn: (r) => r._measurement == \"%s\")", bucket, measurement)
	// Get query client
	queryAPI, err := client1.QueryAPI(org).QueryRaw(context.Background(), query, influxdb2.DefaultDialect())
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	fmt.Println("queryAPI=:", queryAPI)

}

func main2() {

	//连接数据库
	ConnInfluxdb2("yGQHUTBjp1GLhXXP8jXIOhKWOlQcs6dfq4RiwPlawdHCLIpBogMMPuu_UHzrLjfON2m_EnnstZsvHLRUt9DWiQ==", "http://120.27.245.27:8086")

	//measurement string,
	//	tags map[string]string,
	//	fields map[string]interface{},
	//	ts time.Time,
	//

	//生成数据
	//for {

	//cpu Percent
	percent := GetCpuPercent()
	insertCpu := influxdb2.NewPoint("table_Cpu", //measurement string,
		map[string]string{"Cpu": GetIp()},          //	tags map[string]string,
		map[string]interface{}{"percent": percent}, //	fields map[string]interface{},
		time.Now()) //	ts time.Time,
	//插入数据
	WriteInfluxdb2("test", "db01", insertCpu)

	//MEM
	getMem := GetMem()
	insrtMen := influxdb2.NewPoint("table_Men", //表名  //measurement string,
		map[string]string{"men": GetIp()}, //别名 这里列表显示内容 显示为那个ip 的那个指标  tags map[string]string,
		map[string]interface{}{ //	fields map[string]interface{},
			"Total":       getMem.Total,
			"Used":        getMem.Used,
			"Free":        getMem.Free,
			"Available":   getMem.Available,
			"UsedPercent": getMem.UsedPercent,
		}, //展示数据
		time.Now())
	//插入数据
	WriteInfluxdb2("test", "db01", insrtMen)

	//disk
	getDisk := GetDisk()
	for k, v := range getDisk.p {
		insrtDisk := influxdb2.NewPoint("table_Disk", //表名  //measurement string,
			map[string]string{"disk": GetIp() + k}, //别名 这里列表显示内容 显示为那个ip 的那个指标  tags map[string]string,
			map[string]interface{}{ //	fields map[string]interface{},
				"Total":       v.Total,
				"Free":        v.Free,
				"Used":        v.Used,
				"UsedPercent": v.UsedPercent,
			}, //展示数据
			time.Now())

		//插入数据
		WriteInfluxdb2("test", "db01", insrtDisk)
	}

	//查询数据
	//ReadInfluxdb2("test", "db01", "table_Cpu")
	//ReadInfluxdb2("test", "db01", "table_Men")
	//ReadInfluxdb2("test", "db01", "table_Disk")
	time.Sleep(time.Second * 1)
}

//}
