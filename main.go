package main

import (
	zlog "github.com/zeromicro/go-zero/core/logx"
	"github.com/brucewang585/cmplog/util/logx"
	"time"
)

func main() {
	service := "test"
	cfg := zlog.LogConf{
		ServiceName : service,
		Mode : "file",
		Encoding : "json",//           string `json:",default=json,options=[json,plain]"`
		//TimeFormat          string `json:",optional"`
		Path  : "logs/"+service,//              string `json:",default=logs"`
		Level  : "info",//             string `json:",default=info,options=[info,error,severe]"`
		Compress    : false,//        bool   `json:",optional"`
		KeepDays    : 3,//        int    `json:",optional"`
		StackCooldownMillis : 100,//int    `json:",default=100"`
	}
	logx.MustSetup(cfg)
	logx.Info("stop cmpfilesnapshot...\n")

	ttt := "now test 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890"
	t10 := ttt + ttt + ttt + ttt + ttt + ttt + ttt + ttt + ttt
	t100 := t10 + t10 + t10 + t10 + t10 + t10 + t10 + t10 + t10
	t1000 := t100 + t100 + t100 + t100 + t100 + t100 + t100 + t100 + t100

	for i:=0;i<50;i++ {
		logx.Info(t1000)
		time.Sleep(time.Second)
	}

	logx.Close()
}