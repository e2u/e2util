package e2dao

import (
	"fmt"
	"time"

	"strings"

	"git.panda-fintech.com/golang/e2util/e2conf/database/e2model"
	"git.panda-fintech.com/golang/e2util/e2conf/database/e2orm"
	"github.com/jinzhu/gorm"
)

// 通用 DAO 操作包

// Dao 数据库操作
type Dao struct {
	conn *e2orm.Connect
}

// New 建立数据库连接
func New(config *e2orm.Config) *Dao {
	var dao *Dao
	ce := strings.ToLower(config.Endpoint)
	if strings.Contains(ce, "host=") && strings.Contains(ce, "dbname=") {
		dao = &Dao{conn: e2orm.NewPostgreSQL(config)}
	} else {
		dao = &Dao{conn: e2orm.NewMySQL(config)}
	}

	// 注册回调方法
	dao.conn.RW().Callback().Create().Replace("gorm:update_time_stamp", dao.createCallback)
	dao.conn.RW().Callback().Update().Replace("gorm:update_time_stamp", dao.updateCallback)
	return dao
}

// Delete 删除记录
func (d *Dao) Delete(v interface{}, where ...interface{}) error {
	if len(where) > 0 {
		return d.conn.RW().Delete(v, where...).Error
	}
	return d.conn.RW().Delete(v).Error
}

// Exists 檢查記錄指定條件的記錄是否存在
// 從讀寫庫查找，避免數據同步問題
// 使用方法:
// exs := d.Exists(&model.Dictionary{}, "category_code = ? and code = ?", categoryCode, code)
// return exs.Bool, exs.Error
func (d *Dao) Exists(v interface{}, query string, where ...interface{}) *e2model.NullBool {
	var count int
	if err := d.conn.RW(&e2orm.Options{Unscoped: true}).Model(v).
		Where(query, where...).
		Count(&count).Error; err != nil {
		return e2model.NewNullBool(false, err)
	}
	return e2model.NewNullBool(count > 0, nil)
}

// RO 返回数据库只读连接
func (d *Dao) RO(opts ...*e2orm.Options) *gorm.DB {
	return d.conn.RO(opts...)
}

// RW 返回数据库读写连接
func (d *Dao) RW(opts ...*e2orm.Options) *gorm.DB {
	return d.conn.RW(opts...)
}

// Close 关闭数据库连接
func (d *Dao) Close() {
	d.conn.Close()
}

// Save 保存数据
func (d *Dao) Save(v interface{}) error {
	return d.conn.RW().Model(v).Save(v).Error
}

// createCallback 创建新记录回调
func (d *Dao) createCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now().UTC()
		if createdAt, ok := scope.FieldByName("CreatedAt"); ok {
			if createdAt.IsBlank {
				_ = createdAt.Set(nowTime)
			}
		}

		if updatedAt, ok := scope.FieldByName("UpdatedAt"); ok {
			if updatedAt.IsBlank {
				_ = updatedAt.Set(nowTime)
			}
		}
	}
}

// updateCallback 更新记录回调
func (d *Dao) updateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:updated_at"); !ok {
		_ = scope.SetColumn("UpdatedAt", time.Now().UTC())
	}
}

// Patch 只更新部分属性, 用这个方法更新的时候
func (d *Dao) Patch(v interface{}, patchs []*e2model.HttpPatch) error {
	scope := d.conn.RW().NewScope(v)

	updates := make(map[string]interface{})
	for _, patch := range patchs {
		if !scope.HasColumn(patch.Path) {
			return fmt.Errorf("column %v not exists", patch.Path)
		}
		updates[patch.Path] = patch.Value
	}
	return d.conn.RW().Model(v).Update(updates).Error
}
