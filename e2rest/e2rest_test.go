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

var db = e2exec.Must(gorm.Open(postgres.Open("host=pgsql-dev port=5432 user=pgsql password=123456 dbname=e2util_dev sslmode=disable TimeZone=UTC application_name=e2util")))

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
	age, err := e2crypto.RandomNumber[int](15, 40)
	if err != nil {
		t.Fatalf("random age error: %v", err)
	}
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
		updateAge, err := e2crypto.RandomNumber[int](15, 40)
		if err != nil {
			t.Fatalf("random update age error: %v", err)
		}
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