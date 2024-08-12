package e2rest

import (
	"testing"
	"time"

	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2exec"
	"github.com/e2u/e2util/e2test"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db = e2exec.Must(gorm.Open(postgres.Open("host=127.0.0.1 port=5432 user=pgsql password=123456 dbname=e2db_dev sslmode=disable TimeZone=UTC")))

func TestMain(m *testing.M) {
	db.AutoMigrate(Table{})
	m.Run()
}

type Table struct {
	Id        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Age       int
}

func (t *Table) TableName() string {
	return "table_001"
}

func Test_01(t *testing.T) {
	name := e2test.RandomWord()
	age := e2crypto.RandomNumber(15, 40)
	t1 := &Table{
		Name: name,
		Age:  age,
	}
	t.Run("create", func(t *testing.T) {
		r, err := Create(db, t1)
		if err != nil {
			t.Fatal(err)
		}
		if r.Name != name {
			t.Fatal("name not match")
		}
		t.Log(r)
	})

	t.Run("update", func(t *testing.T) {
		updateName := e2test.RandomWord()
		updateAge := e2crypto.RandomNumber(15, 40)
		t1.Name = updateName
		t1.Age = updateAge
		r, err := Update(db, t1)
		if err != nil {
			t.Fatal(err)
		}
		if r.Name != updateName {
			t.Fatal("name not match")
		}
		t.Log(r)
	})

}
