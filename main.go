package main

import (
	"context"
	"flag"
	"fmt"

	"os"
	"time"

	"github.com/brucewang585/cmplog/util/logx"
	zlog "github.com/zeromicro/go-zero/core/logx"
	ztrace "github.com/brucewang585/cmplog/trace"

	//"go.opentelemetry.io/otel"
	//"go.opentelemetry.io/otel/trace"
	//"go.opentelemetry.io/otel/attribute"
	//"github.com/open-telemetry/opentelemetry-go"
	//"github.com/open-telemetry/opentelemetry-go/attribute"
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

	tp := ztrace.NewTracerProvider()
	tracer := tp.Tracer("ex.com/webserver")

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

			ctx, sp := tracer.Start(context.Background(), "parent")
			//event := fmt.Sprintf("Now: %s", t.Format("2006.01.02 15:04:05"))
			//span.AddEvent(event, trace.WithAttributes(label.Int("bogons", 100)))
			//span.SetAttributes(anotherKey.String("yes"))
			sp.End()

			id += 1
			dd,_ := sp.SpanContext().MarshalJSON()
			fl.WithDuration(time.Second).WithContext(ctx).Infof("id:%09d, test trace-parent,info:%s",id,string(dd))

			ctx, sc := tracer.Start(ctx, "child")
			sc.End()
			id += 1
			dd,_ = sc.SpanContext().MarshalJSON()
			fl.WithDuration(time.Second).WithContext(ctx).Infof("id:%09d, test trace-child,info:%s",id,string(dd))
		}
	}

	fl.Close()
}