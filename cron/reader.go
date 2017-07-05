package cron

import (
	"encoding/json"
	"time"

	"github.com/ZeaLoVe/alarm-assist/g"
	"github.com/astaxie/beego"
	"github.com/garyburd/redigo/redis"
	"github.com/open-falcon/common/model"
)

func ComsumeEvent() {
	for i := 0; i < g.Config().Redis.MaxConsumer; i++ {
		go ReadEvent()
	}
}

func ReadEvent() {
	queues := g.Config().Redis.QueryQueues
	if len(queues) == 0 {
		return
	}
	time.Sleep(3 * time.Second)

	for {
		event, err := popEvent(queues)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		consume(event)
	}
}

func popEvent(queues []string) (*model.Event, error) {

	count := len(queues)

	params := make([]interface{}, count+1)
	for i := 0; i < count; i++ {
		params[i] = queues[i]
	}
	// set timeout 0
	params[count] = 0

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	reply, err := redis.Strings(rc.Do("BRPOP", params...))
	if err != nil {
		beego.Debug("get alarm event from redis fail: %v", err)
		return nil, err
	}

	var event model.Event
	err = json.Unmarshal([]byte(reply[1]), &event)
	if err != nil {
		beego.Debug("parse alarm event fail: %v", err)
		return nil, err
	}

	beego.Debug("======>>>>")
	beego.Debug(event.String())

	g.Events.Put(&event)

	return &event, nil
}
