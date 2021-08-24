#!/etc/bin

echo "hello"
# 启动抓包程序
echo `date "+%Y-%m-%d %H:%M:%S"`
echo "启动抓包程序..."
echo `go run main.go >/dev/null 2>&1 &`
# 启动发包程序
echo `date "+%Y-%m-%d %H:%M:%S"`
echo "启动发包程序..."
echo `./analyze_db -H 9.145.42.250 -P 3306 -u analyze_rw -p analyze_password -d analyze_db  -t analyze_002 >/dev/null 2>&1 &`
# 同步上报数据
echo `date "+%Y-%m-%d %H:%M:%S"`
echo "开始上报数据程序..."

i=15
while :
do
	# 每分钟到15秒就上报一次
	now=`date +%S`
	if [ $now == $i ]
	then
		echo `date "+%Y-%m-%d %H:%M:%S"`
		# 调用input_go.go查看当前时刻前一分钟内的各指令增量
		# 输出格式为"指令类型 增量",重定向作为参数传入upload_data_toZhiyan.py完成一次上报
		echo `go run input_go.go | xargs python upload_data_toZhiyan.py`
		
		# 睡眠1秒钟，防止当前分钟秒数重复上报
		sleep 1s
	fi

done


