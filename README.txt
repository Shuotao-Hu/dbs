=======   入口   ========
》执行方法
source ./loadToZhiyan.sh
》loadToZhiyan.sh
并行执行抓包程序、发包程序、数据同步上报智研，三个操作

======= 程序说明 ========
- main.go 
抓包程序

- analyze_db
发包程序

- input_go.go
获取当前时刻前一分钟内的各指令增量

- upload_data_toZhiyan.py
将输入指令类型、增量参数上报智研监控宝

- loadToZhiyan.sh
并行执行抓包程序、发包程序、数据同步上报智研，三个操作
