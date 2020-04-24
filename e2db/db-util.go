package e2db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func writeRows(w io.Writer, rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			continue
		}

		record := make(map[string]interface{})
		for i, col := range values {
			if col == nil {
				continue
			}
			switch t := col.(type) {
			default:
				return fmt.Errorf("Unexpected type %T\n", t)
			case bool:
				record[columns[i]] = col.(bool)
			case int:
				record[columns[i]] = col.(int)
			case int64:
				record[columns[i]] = col.(int64)
			case float64:
				record[columns[i]] = col.(float64)
			case string:
				record[columns[i]] = col.(string)
			case time.Time:
				record[columns[i]] = time.Time(col.(time.Time)).Format("2006-01-02 15:04:05.999999999Z07:00")
			case []byte: // -- all cases go HERE!
				record[columns[i]] = string(col.([]byte))
			}
		}
		s, err := json.Marshal(record)
		if err != nil {
			return err
		}
		_, _ = w.Write(s)
		_, _ = io.WriteString(w, "\n")
	}
	return nil
}

// DumpTable2JSON 根据  sql 查询语句 dump 数据成 json ,注意,目前这个版本只支持 postgreSQL 的分页模式
func DumpTable2JSON(db *sql.DB, w io.Writer, sql string) error {
	totalCountRow := db.QueryRow(fmt.Sprintf("SELECT COUNT(1) FROM (%s) AS Z", sql))
	totalCount := 0
	if err := totalCountRow.Scan(&totalCount); err != nil {
		return err
	}
	batchMax := 20000
	for i := 0; i <= (totalCount/batchMax)+1; i++ {
		offset := i * batchMax
		rows, err := db.Query(fmt.Sprintf("%s LIMIT %d OFFSET %d;", sql, batchMax, offset))
		if err != nil {
			return err
		}
		if err := writeRows(w, rows); err != nil {
			return err
		}
		rows.Close()
	}
	return nil
}

// PgConnection 建立 PostgreSQL 数据库连接
func PgConnection(host string, port int, username, password, dbname string, ssl bool) (*gorm.DB, error) {
	sslmode := "disable"
	if ssl {
		sslmode = "require"
	}
	return gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, username, password, dbname, sslmode))
}

// MySQLConnection 建立 MySQL 数据库连接
func MySQLConnection(host string, port int, username, password, dbname string) (*gorm.DB, error) {
	return gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, dbname))
}

// Connection 建立数据库连接
func Connection(dialect, host string, port int, username, password, dbname string, ssl bool) (*gorm.DB, error) {
	switch dialect {
	case "postgres", "pgsql":
		return PgConnection(host, port, username, password, dbname, ssl)
	case "mysql":
		return MySQLConnection(host, port, username, password, dbname)
	}
	return nil, fmt.Errorf("unknow dialect")
}
