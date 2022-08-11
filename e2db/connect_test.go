package e2db

import (
	"testing"
	"time"
)

func TestNew_try_server_down(t *testing.T) {

	dsn := "host=127.0.0.1 port=5432 user=postgres password=none dbname=test application_name=unit-test"
	cfg := &Config{
		Writer: dsn,
	}
	conn := New(cfg)
	const maxTick = 120
	var runCount = 0
	var rs string

	for tick := range time.Tick(1 * time.Second) {
		runCount += 1
		if runCount >= maxTick {
			break
		}
		if err := conn.RW().Raw("SELECT NOW()").Scan(&rs).Error; err != nil {
			t.Errorf("db conn error=%v", err)
			continue
		}
		t.Logf("[#%04d] tick=%v, result=%v", runCount, tick, rs)
	}

}
