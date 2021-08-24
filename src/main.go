/*
	程序使用说明：
	1. 使用之前：需要执行以下命令安装依赖包
		go get -u github.com/go-sql-driver/mysql
		go get git@github.com:google/gopacket.git
	2. 功能分包信息
		（1）capture包：
			startCapture：网络抓包主函数
			getFilter：定义过滤器
			subStr：截取可能含有非法字符的字符串
		（2）pkgg包：

 */
package main

import (
	"fmt"
	"strconv"
	"time"


	"database/sql"


	_ "github.com/go-sql-driver/mysql"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const TABLENAME = "analyze_result_002" // 声明静态变量（输出表名）

func main() {
	// 1.连接远程数据库
	fmt.Println("Connect to database:analyze_db...")
	db,errConnect := connectDB()
	if errConnect != nil {
		fmt.Println(errConnect)
	}
	// 2.判断目标数据表是否建立，若无则按照规则建立数据表
	CreateTable(db)

	// 3.清空数据表已有数据
	DeleteData(db)

	// 4.开始运行抓包程序
	fmt.Println("Capture start..")
	startCapture(db)

}

//定义过滤器
func getFilter(port uint16) string {
	filter := fmt.Sprintf("tcp and ((src port %v) or (dst port %v))",  port+1, port)
	return filter
}


//截取字符串 start 起点下标 end 终点下标(不包括)
func subStr(str string, start int, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		return ""
	}

	if end < 0 || end > length {
		return ""
	}
	return string(rs[start:end])
}
// 监控抓包函数，输出包文件为pcap文件
func startCapture(db *sql.DB) {
	deviceName := "eth1" // 设备名
	snapLen := int32(65535)
	port := uint16(3306)
	filter := getFilter(port) // 设置过滤器
	fmt.Printf("device:%v, snapLen:%v, port:%v\n", deviceName, snapLen, port)
	fmt.Println("filter:", filter)

	//打开网络接口，抓取在线数据
	handle, err := pcap.OpenLive(deviceName, snapLen, true, pcap.BlockForever)
	if err != nil {
		fmt.Printf("pcap open live failed: %v", err)
		return
	}

	// 设置过滤器
	if err := handle.SetBPFFilter(filter); err != nil {
		fmt.Printf("set bpf filter failed: %v", err)
		return
	}
	// 抓包
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.NoCopy = true

	// 声明captures为CaptuRes类型,用于封装格式化的数据
	var captures CaptuRes

	//检查抓到的网络包是否合法，如合法则接收
	for packet := range packetSource.Packets() {

		// 判断抓回的网络包是否合法
		if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
			fmt.Println("Unexpected packet")
			// 直接进行下一次抓包
			continue
		} else {
			//利用GC机制回收旧结构体并创建新结构体
			captures = CaptuRes{}

			//进行pcap抓包分析
			packetAnalyse(db,packet,captures)
		}
	}
}

// 对pcap进行抓包分析，包括关键信息提取
func packetAnalyse(db *sql.DB,packet gopacket.Packet,captures CaptuRes) {
	// 提取tcp层进行数据库命令提取
	tcp := packet.TransportLayer().(*layers.TCP)
	payload := fmt.Sprintf("%s", tcp.Payload)

	// 提取IPv4层进行其他参数的提取
	ipLayer := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)


	if len(payload) >= 0 {
		// 截取命令头以便过滤
		payLoadHead := subStr(payload, 5, 11)
		if payLoadHead == "SELECT" || payLoadHead == "DELETE" || payLoadHead == "UPDATE" || payLoadHead == "INSERT" {

			// 该请求包的命令
			captures.sqlOrder = fmt.Sprintf(subStr(payload, 5, len(payload)-1))
			//fmt.Println(captures.sqlOrder)

			// 该请求包的命令类型
			captures.sqlType = payLoadHead
			//fmt.Println(captures.sqlType)

			// 该请求的包时间
			t := packet.Metadata().CaptureInfo.Timestamp
			captures.timeStamp = fmt.Sprintf("%d-%d-%d %d:%d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
			//fmt.Println(captures.timeStamp)

			// 该请求包的目标ip
			captures.dstip = fmt.Sprintf("%s", ipLayer.DstIP.String())
			//fmt.Println(captures.dstip)

			// 该请求包采用的协议
			captures.protocol = fmt.Sprintf("%s", ipLayer.Protocol) // 将该条结果导入数据库
			//fmt.Println(captures.protocol)

			// 将抓包结果导入目标数据库
			go exportToDB(db, captures)
		}
	}
}

// connect mysql database
func connectDB()(db *sql.DB,err error) {
	//进行远程数据库连接
	db,openErr := sql.Open("mysql","analyze_rw:analyze_password@tcp(9.145.42.250:3306)/analyze_db")
	if openErr != nil {
		err = openErr
		return nil,err
	}
	// 验证远程数据库的连接状况
	if pingErr := db.Ping(); pingErr != nil {
		fmt.Println("Fail to open analyze_db！")
		err = pingErr
		return nil,err
	}
	fmt.Println("Connect to analyze_db seccessfully!")
	return db,err
}

// create table if not exists
func CreateTable(db *sql.DB){
	//create table : analyze_result_02
	//del_sql := "DROP TABLE IF EXISTS analyze_result_02;"
	tableSql :=
		"CREATE TABLE IF NOT EXISTS "+ TABLENAME +"("+
			"`id` BIGINT(20) NOT NULL AUTO_INCREMENT comment '编号',"+
			"`time` timestamp not null default current_timestamp comment '写入时间',"+
			"`ip` varchar(15) not null default '192.168.0.0' comment '目标IP',"+
			"`protocol` char(3) not null default 'TCP' comment '协议',"+
			"`command` varchar(500) not null default 'select * from table;' comment '命令',"+
			"`type_command` varchar(10) not null default 'select' comment '命令类型',"+
			"PRIMARY KEY(id)"+
			")ENGINE=InnoDB DEFAULT CHARSET=utf8 comment '抓包结果统计表';"
	if db == nil {
		fmt.Println("db is nil!")
	}
	res, err := db.Exec(tableSql)
	if err != nil{
		fmt.Printf("create table failed, err:[%v]", err.Error())
		return
	}
	aff, _ := res.RowsAffected()
	fmt.Printf("create table success, affected rows:[%d]\n", aff)
	return
}

// 导出至远程数据库
func exportToDB(db *sql.DB,captures CaptuRes) {
	_, err := db.Exec("insert into "+ TABLENAME +"(id,time,ip,protocol,command,type_command) values(default,?,?,?,?,?)",
		captures.timeStamp,captures.dstip,captures.protocol,captures.sqlOrder,captures.sqlType)
	if err != nil {
		fmt.Printf("Insert data failed, error:[%v]\n", err.Error())
	}
}

// insert data
func InsertData(db *sql.DB) {
	var affTotal int64 = 0
	for i:=0; i<20; i++ {
		res, err := db.Exec("insert into "+ TABLENAME +"(id,time,ip,protocol,command,type_command) values(default,default,default,default,default,'select')")
		if err != nil {
			fmt.Printf("insert data failed, error:[%v]", err.Error())
			return
		}
		aff, _ := res.RowsAffected()
		affTotal += aff
	}
	fmt.Printf("insert data success, affected rows:[%d]\n", affTotal)
	return
}

// count total nums
func TotalData(db *sql.DB) int {
	totalData := "select count(id) from " + TABLENAME + ";"
	rows, err := db.Query(totalData)
	if err != nil{
		fmt.Printf("count data failed, err:[%v]", err.Error())
		return -1
	}
	defer rows.Close()
	var param string
	for rows.Next(){
		err := rows.Scan(&param)
		if err != nil{
			fmt.Println(err)
			return -1
		}
	}
	paramInt, err := strconv.Atoi(param)
	return paramInt
}

// delete data
func DeleteData(db *sql.DB){
	deleteData := "delete from " + TABLENAME + ";"
	res, err := db.Exec(deleteData)
	if err != nil{
		fmt.Printf("delete data failed, err:[%v]", err.Error())
		return
	}
	aff, _ := res.RowsAffected()
	fmt.Printf("delete data success, affected rows:[%d]\n", aff)
	return
}


// count nums of type_command
func Statistics (db *sql.DB) map[string]string {
	timeNow := time.Now()
	itl, _ := time.ParseDuration("-1m")   //time interval 1 min
	timeBegin := timeNow.Add(itl)
	timeNowF := timeNow.Format("2006-01-02 15:04:05")
	timeBeginF := timeBegin.Format("2006-01-02 15:04:05")
	staSql := "select type_command, count(type_command) as nums from "+ TABLENAME +" where time between '"+ timeBeginF +"' and '"+ timeNowF +"' group by type_command;"
	//staSql := "select type_command, count(type_command) as nums from "+ tableName +" group by type_command;"
	rows, err := db.Query(staSql)
	if err != nil{
		fmt.Printf("statistics failed, err:[%v]", err.Error())
		return nil
	}
	defer rows.Close()
	ansMap := make(map[string]string)
	for rows.Next(){
		var param1,param2 string
		err := rows.Scan(&param1, &param2)
		if err != nil{
			fmt.Println(err)
			return nil
		}
		ansMap[param1] = param2
	}
	return ansMap
}
