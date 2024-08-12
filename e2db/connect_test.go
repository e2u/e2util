package e2db

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/e2u/e2util/e2regexp"
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

func Test_parseMySQLdsn(t *testing.T) {
	// phoneRENamedCaps := `(?P<area>\d{3})\-(?P<exchange>\d{3})\-(?P<line>\d{4})$`
	dsn := "root:password@tcp(192.168.10.29:3306)/mycash_test?charset=utf8mb4&parseTime=True&loc=Local"
	re := regexp.MustCompile(`^(?P<userinfo>[^@]+)@(?P<conn>[^/]+)/(?P<dbname>[^\?]+)\?(?P<params>.+)$`)
	rs, ok := e2regexp.NamedFindStringSubmatch(dsn, re)
	fmt.Println(rs, ok)
	fmt.Println(rs["userinfo"])
	fmt.Println(rs["conn"])
	fmt.Println(rs["dbname"])
	fmt.Println(rs["params"])

}
