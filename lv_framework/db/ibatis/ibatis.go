// Package dotsql provides a way to separate your code from SQL Queries.
//
// It is not an ORM, it is not a query builder.
// Dotsql is a library that helps you keep sql files in one place and use it with ease.
//
// For more usage examples see https://github.com/qustavo/dotsql
package ibatis

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/morrisxyang/xreflect"
	"github.com/spf13/cast"
	"io"
	"lostvip.com/logme"
	"lostvip.com/utils/lib_file"
	"os"
	"reflect"
	"robvi/app/global"
	"strings"
)

// Execer is an interface used by Exec.
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// ExecerContext is an interface used by ExecContext.
type ExecerContext interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// IBatis represents a dotSQL Queries holder.
type IBatis struct {
	Queries     map[string]string
	TplFile     string
	CurrBaseSql string
}

/**
 * 从mapper目录解析sql文件
 */
func NewIBatis(relativePath string) (string *IBatis, err error) {
	basePath, _ := os.Getwd()
	absolutePath := basePath + "/mapper" //为了方便管理，必须把映射文件放到mapper目录
	if strings.HasPrefix(relativePath, "/") {
		absolutePath = absolutePath + relativePath
	} else {
		absolutePath = absolutePath + "/" + relativePath
	}
	dot, err := LoadFromFile(absolutePath)
	if err != nil {
		panic(err)
	}
	return dot, err
}

func (e *IBatis) GetSql(tagName string, params interface{}) (string, error) {
	cfg := global.GetConfigInstance()
	query, err := e.LookupQuery(tagName)
	if err != nil || query == "" {
		panic("tpl文件格式错误!")
	}
	if cfg.IsDebug() {
		logme.Log.Info("========>SQL:  " + query)
	}
	//动态解析
	sql, err := lib_file.ParseTemplateStr(query, params)
	if err != nil {
		panic(err)
	}
	if sql == "" {
		panic("sql文件存在语法错误，请使用golang的telmplate标准语法" + e.getTplFile())
	}
	e.CurrBaseSql = sql //缓存当前正在执行的分页sql
	return sql, err
}

/**
 * 从mapper目录解析sql文件
 */
func (e *IBatis) GetLimitSql(tagName string, params interface{}) string {
	sql, err := e.GetSql(tagName, params)
	if err != nil {
		panic(err)
	}
	paramType := reflect.TypeOf(params).Kind()
	if paramType == reflect.Map {
		paramMap := params.(map[string]interface{})
		pageNum := paramMap["pageNum"]
		pageSize := paramMap["pageSize"]
		if pageSize == nil || pageNum == nil {
			panic("???????????分页信息错误" + sql)
		} else {
			start := cast.ToInt64(pageSize) * (cast.ToInt64(pageNum) - 1)
			sql = sql + " limit  " + cast.ToString(start) + "," + cast.ToString(pageNum)
		}
	} else {
		pageNum, err1 := xreflect.FieldValue(params, "PageNum")
		pageSize, err2 := xreflect.FieldValue(params, "PageSize")
		if pageSize == 0 || pageNum == 0 || err1 != nil || err2 != nil {
			panic("???????????分页信息错误" + sql)
		} else {
			start := cast.ToInt64(pageSize) * (cast.ToInt64(pageNum) - 1)
			sql = sql + " limit  " + cast.ToString(start) + "," + cast.ToString(pageNum)
		}
	}

	return sql
}

func (e *IBatis) GetCountSql() string {
	cfg := global.GetConfigInstance()
	//动态解析

	if e.CurrBaseSql == "" {
		panic("未初始化过ibatis对象,未传入过sql参数！" + e.getTplFile())
	}
	sql := " select count(1) from (" + e.CurrBaseSql + ") t "
	if cfg.IsDebug() {
		logme.Log.Info("========>SQL: \n " + sql)
	}
	return sql
}

func (d IBatis) LookupQuery(name string) (query string, err error) {
	query, ok := d.Queries[name]
	if !ok {
		err = fmt.Errorf("dotsql: '%s' could not be found", name)
	}

	return
}

// Exec is a wrapper for database/sql's Exec(), using dotsql named query.
func (d IBatis) Exec(db Execer, name string, args ...interface{}) (sql.Result, error) {
	query, err := d.LookupQuery(name)
	if err != nil {
		return nil, err
	}

	return db.Exec(query, args...)
}

// ExecContext is a wrapper for database/sql's ExecContext(), using dotsql named query.
func (d IBatis) ExecContext(ctx context.Context, db ExecerContext, name string, args ...interface{}) (sql.Result, error) {
	query, err := d.LookupQuery(name)
	if err != nil {
		return nil, err
	}

	return db.ExecContext(ctx, query, args...)
}

// GetRawSql returns the query, everything after the --name tag
func (d IBatis) GetRawSql(name string) (string, error) {
	return d.LookupQuery(name)
}

// GetQueryMap returns a map[string]string of loaded Queries
func (d IBatis) GetQueryMap() map[string]string {
	return d.Queries
}

func (e *IBatis) getTplFile() string {
	return e.TplFile
}

// Load imports sql Queries from any io.Reader.
func Load(r io.Reader) (*IBatis, error) {
	scanner := &Scanner{}
	queries := scanner.Run(bufio.NewScanner(r))

	dotsql := &IBatis{
		Queries: queries,
	}

	return dotsql, nil
}

// LoadFromFile imports SQL Queries from the file.
func LoadFromFile(sqlFile string) (*IBatis, error) {
	f, err := os.Open(sqlFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Load(f)
}

// LoadFromString imports SQL Queries from the string.
func LoadFromString(sql string) (*IBatis, error) {
	buf := bytes.NewBufferString(sql)
	return Load(buf)
}

// Merge takes one or more *IBatis and merge its Queries
// It's in-order, so the last source will override Queries with the same name
// in the previous arguments if any.
func Merge(dots ...*IBatis) *IBatis {
	queries := make(map[string]string)

	for _, dot := range dots {
		for k, v := range dot.GetQueryMap() {
			queries[k] = v
		}
	}

	return &IBatis{
		Queries: queries,
	}
}
