package main

import (
	"fmt"
	"time"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	
)



func main() {
	// 1.连接远程数据库
	db,errConnect := connectDB()
	if errConnect != nil {
		fmt.Println(errConnect)
		Statistics(db)
	}
	Statistics(db)
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
	//fmt.Println("Connect to analyze_db seccessfully!")
	return db,err
}

// 为了将该函数的输出作为上报程序的输入，这里将返回的map注释掉
// 改用输出结果，重定向到上报程序
// count nums of type_command
func Statistics (db *sql.DB) {
	TABLENAME:="analyze_result_002"
	timeNow := time.Now()
	
	itl, _ := time.ParseDuration("-1m")   //time interval 1 min
	timeBegin := timeNow.Add(itl)
	timeNowF := timeNow.Format("2006-01-02 15:04:05")	
	timeBeginF := timeBegin.Format("2006-01-02 15:04:05")
	staSql := "select type_command, count(type_command) as nums from "+ TABLENAME +" where time between '"+ timeBeginF +"' and '"+ timeNowF +"' group by type_command;"
	rows, err := db.Query(staSql)
	if err != nil{
		fmt.Printf("statistics failed, err:[%v]", err.Error())
		//return nil
	}
	defer rows.Close()
	
	//ansMap := make(map[string]string)
	//fmt.Println("In...1")
	for rows.Next(){
		var param1,param2 string
		rows.Scan(&param1, &param2)
		fmt.Println(param1+ " " +param2)
		
		//if err != nil{
		//	fmt.Println(err)
		//	return nild
		//}
		//ansMap[param1] = param2
	}
	
}



