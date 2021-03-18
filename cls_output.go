package main

import (
	"time"

	"github.com/golang/glog"
	proto "github.com/huangjacky/gohangout-output-cls/cls"
	cls "github.com/huangjacky/gohangout-output-cls/proto"
)

// ClsOutput 插件
type ClsOutput struct {
	config    map[interface{}]interface{}
	Region    string
	SecretId  string
	SecretKey string
	Token     string
	Inner     bool
	Logset    string
	Topic     string
	MaxBytes  int
	MaxSize   int
	BufLength int
	Tick      int
	channel   chan *proto.Log
	client    *cls.Client
}

/*
New 插件模式的初始化
*/
func New(config map[interface{}]interface{}) interface{} {
	p := &ClsOutput{
		config:  config,
		channel: make(chan *proto.Log),
	}
	if v, ok := config["region"]; ok {
		p.Region = v.(string)
	} else {
		glog.Fatal("region is unset")
	}
	if v, ok := config["sid"]; ok {
		p.SecretId = v.(string)
	} else {
		glog.Fatal("sid is unset")
	}
	if v, ok := config["skey"]; ok {
		p.SecretKey = v.(string)
	} else {
		glog.Fatal("skey is unset")
	}
	if v, ok := config["token"]; ok {
		p.Token = v.(string)
	} else {
		glog.Fatal("token is unset")
	}

	if v, ok := config["logset"]; ok {
		p.Logset = v.(string)
	} else {
		glog.Fatal("logset is unset")
	}
	if v, ok := config["topic"]; ok {
		p.Topic = v.(string)
	} else {
		glog.Fatal("topic is unset")
	}
	if v, ok := config["inner"]; ok {
		p.Inner = v.(bool)
	} else {
		p.Inner = true
	}
	if v, ok := config["max_bytes"]; ok {
		p.MaxBytes = v.(int)
	} else {
		p.MaxBytes = 1 * 1024 * 1024
	}

	if v, ok := config["max_size"]; ok {
		p.MaxSize = v.(int)
	} else {
		p.MaxSize = 1024
	}
	if v, ok := config["tick"]; ok {
		p.Tick = v.(int)
	} else {
		p.Tick = 4
	}
	tick := time.NewTicker(time.Duration(p.Tick) * time.Second)
	go func() {
		count = 0
		bytesCount = 0

		for {
			select {
			case dd, ok := <-d:
				if ok {
					count++
					bytesCount += dd.length
					events = append(events, dd.msg)
				}
			case <-tick.C:
				send = true
			}
			if send || count >= m.Output.eventCount || bytesCount >= m.Output.bytesCount {
				send = false
				if count > 0 {
					m.Output.send2CLS(topicId, events)
					reset()
				}
			}
		}

	}()
	var inet Net
	if p.Inner {
		inet = InNet
	} else {
		inet = OutNet
	}
	p.client = NewClient(
		p.Region, p.SecretId, p.SecretKey, p.Token, inet,
	)
	p.BufLength = 0
	return p
}

//Emit 单次事件的处理函数
func (p *ClsOutput) Emit(event map[string]interface{}) {

}

//Shutdown 关闭需要做的事情
func (p *ClsOutput) Shutdown() {

}
