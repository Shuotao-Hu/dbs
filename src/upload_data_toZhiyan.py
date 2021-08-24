import urllib.request
import json
import random
import sys

def main(argv):
    #智研上报的URL
    url='http://zhiyan.monitor.access.inner.woa.com:8080/access_v1.http_service/HttpCurveReportRpc'
    report_data = [
        {"metric":str(argv[1]),"value": argv[2],"tags":{"metricType": "test"}},
        {"metric":str(argv[3]),"value": argv[4],"tags":{"metricType": "test"}},
        {"metric":str(argv[5]),"value": argv[6],"tags":{"metricType": "test"}},
        {"metric":str(argv[7]),"value": argv[8],"tags":{"metricType": "test"}}]
    report_datas = json.dumps(report_data)
    rand_num = random.randint(100000, 999999)
    request_data = {
        "pkg_seq": rand_num,
        "report_cnt": 1 ,
        "app_mark": "3633_20699_test_qinglin",
        "sec_lvl_en_name": "default",
        "env":"prod",
        "report_data": report_datas,
        "instance_mark": "9.134.235.217",
        "report_ip": "9.134.235.217",
        "data_type": 0 ,
        "method": 1
    }

    report_datas = json.dumps(request_data)
    headers={"Content-Type":"application/json"}
    report_datas=bytes(report_datas,'utf8')
    #发送POST请求
    request = urllib.request.Request(url, report_datas, headers)
    response = urllib.request.urlopen(request)
    ret = response.read()
    print (ret)
    
if __name__ == "__main__":
    main(sys.argv)
