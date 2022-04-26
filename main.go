package main

import (
	"flag"
	"fmt"

	"github.com/brucewang585/cmplog/util/logx"
	//"github.com/opentracing/opentracing-go"
	zlog "github.com/zeromicro/go-zero/core/logx"
	//"go.opentelemetry.io/otel/trace"
	"os"
	"time"

	//"github.com/brucewang585/cmplog/mocktracer"
)

var (
	service string
)
func init() {
	flag.StringVar(&service,"s","test","service name")
	flag.Parse()

	if service == "" {
		fmt.Println("service is empty")
		os.Exit(0)
	}
}

func main() {
	go testLog(service)
	testFullLog(service+"_full")
}

func testLog(service string) {
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

	tm := time.NewTicker(time.Second)
	defer tm.Stop()

	id := 0
	for {
		select {
		case <- tm.C:
			id += 1
			logx.Infof("id:%09d, test info",id)

			id += 1
			logx.Errorf("id:%09d, test error",id)
		}
	}

	logx.Close()
}

func testFullLog(service string) {
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
	fl := logx.NewFullLogger(cfg)

	tm := time.NewTicker(time.Second)
	defer tm.Stop()

	id := 0
	for {
		select {
		case <- tm.C:
			id += 1
			fl.Infof("id:%09d, test info",id)

			id += 1
			fl.Errorf("id:%09d, test error",id)

			id += 1
			fl.WithDuration(time.Second).Slowf("id:%09d, test slow",id)

			//id += 1
			//fl.Severef("id:%09d, test severe",id)

			/*
			id += 1
			span1 := mocktracer.TTracer.StartSpan(
				"1",
				opentracing.Tags(map[string]interface{}{"x": "y"}))

			span2 := span1.Tracer().StartSpan(
				"1.1", opentracing.ChildOf(span1.Context()))
			span2.Finish()
			span1.Finish()

			context.Background()

			fl.WithDuration(time.Second).WithContext(trace.ContextWithSpan(context.Background(),span1))
			fl.WithDuration(time.Second).WithContext(trace.ContextWithSpan(span2))
			*/
		}
	}

	fl.Close()
}