package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"strconv"
	"time"
)

// connect to database
func connect()(db *sqlx.DB){
	database, err := sqlx.Open("mysql", "analyze_rw:analyze_password@tcp(9.145.42.250:3306)/analyze_db?charset=utf8")
	if err != nil {
		fmt.Println("open mysql failed,", err)
		return
	}
	db = database
	fmt.Println("open mysql success.")
	return db
}

// create table if not exists
func CreateTable(db *sqlx.DB){
	//create table : analyze_result_02
	//del_sql := "DROP TABLE IF EXISTS analyze_result_02;"
	tableSql :=
		"CREATE TABLE IF NOT EXISTS "+ tableName +"("+
			"`id` BIGINT(20) NOT NULL AUTO_INCREMENT comment '编号',"+
			"`time` timestamp not null default current_timestamp comment '写入时间',"+
			"`ip` varchar(15) not null default '192.168.0.0' comment '目标IP',"+
			"`protocol` char(3) not null default 'TCP' comment '协议',"+
			"`command` varchar(500) not null default 'select * from table;' comment '命令',"+
			"`type_command` varchar(10) not null default 'select' comment '命令类型',"+
			"PRIMARY KEY(id)"+
			")ENGINE=InnoDB DEFAULT CHARSET=utf8 comment '抓包结果统计表';"
	res, err := db.Exec(tableSql)
	if err != nil{
		fmt.Printf("create table failed, err:[%v]", err.Error())
		return
	}
	aff, _ := res.RowsAffected()
	fmt.Printf("create table success, affected rows:[%d]\n", aff)
	return
}

// insert data
func InsertData(db *sqlx.DB) {
	var affTotal int64 = 0
	for i:=0; i<20; i++ {
		res, err := db.Exec("insert into "+ tableName +"(id,time,ip,protocol,command,type_command) values(default,default,default,default,default,'select')")
		if err != nil {
			fmt.Printf("insert data faild, error:[%v]", err.Error())
			return
		}
		aff, _ := res.RowsAffected()
		affTotal += aff
	}
	fmt.Printf("insert data success, affected rows:[%d]\n", affTotal)
	return
}

// count total nums
func TotalData(db *sqlx.DB) int {
	totalData := "select count(id) from " + tableName + ";"
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
func DeleteData(db *sqlx.DB){
	deleteData := "delete from " + tableName + ";"
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
func Statistics (db *sqlx.DB) map[string]string {
	timeNow := time.Now()
	itl, _ := time.ParseDuration("-1m")   //time interval 1 min
	timeBegin := timeNow.Add(itl)
	timeNowF := timeNow.Format("2006-01-02 15:04:05")
	timeBeginF := timeBegin.Format("2006-01-02 15:04:05")
	staSql := "select type_command, count(type_command) as nums from "+ tableName +" where time between '"+ timeBeginF +"' and '"+ timeNowF +"' group by type_command;"
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