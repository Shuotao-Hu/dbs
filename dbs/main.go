package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

//command: mysql -h9.145.42.250 -P3306 -uanalyze_rw -panalyze_password analyze_db

var tableName string = "test_02" // 库表名称

func main(){
	var db *sqlx.DB = connect()
	defer db.Close()
	CreateTable(db)
	InsertData(db)
	ansMap :=  Statistics(db)    // 统计一分钟内的命令类型数，返回map
	fmt.Println("type_command nums:")
	for key, value := range ansMap {
		fmt.Println(key,value)
	}
	totalData := TotalData(db) // 当库表中数据多于10000条时清空库表
	fmt.Printf("total data: %d\n",totalData)
	if totalData > 10000 {
		DeleteData(db)
	}
}
