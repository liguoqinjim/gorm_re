package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"strings"
)

type Config struct {
	DBHost string
	DBUser string
	DBPwd  string
	DBName string
}

var config Config

func LoadConfig(data []byte) {
	err := json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	if config.DBName == "DBName" {
		log.Fatal("请修改配置文件中的数据库配置")
	}
}

type Column struct {
	TableCataLog           sql.NullString
	TableSchema            sql.NullString
	TableName              sql.NullString
	ColumnName             sql.NullString
	OrdinalPosition        sql.NullString
	ColumnDefault          sql.NullString
	IsNullable             sql.NullString
	DataType               sql.NullString
	CharacterMaximumLength sql.NullString
	CharacterOctetLength   sql.NullString
	NumericPrecision       sql.NullString
	NumericScale           sql.NullString
	CharacterSetName       sql.NullString
	CollationName          sql.NullString
	ColumnType             sql.NullString
	ColumnKey              sql.NullString
	Extra                  sql.NullString
	Privileges             sql.NullString
	ColumnComment          sql.NullString
}

func GetColumns() []*Column {
	connectInfo := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8&parseTime=True&loc=Local", config.DBUser, config.DBPwd, config.DBHost, "information_schema")
	db, err := sql.Open("mysql", connectInfo)
	if err != nil {
		log.Fatal(err)
	}

	//ping
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//select
	querySql := "select * from `COLUMNS` where TABLE_SCHEMA = ? order by TABLE_NAME,ORDINAL_POSITION"
	rows, err := db.Query(querySql, config.DBName)
	if err != nil {
		log.Fatal(err)
	}

	//columnNames, err := rows.Columns()
	//if err != nil {
	//	log.Fatal(err)
	//}

	columns := make([]*Column, 0)

	for rows.Next() {
		column := new(Column)
		err = rows.Scan(&column.TableCataLog, &column.TableSchema, &column.TableName, &column.ColumnName, &column.OrdinalPosition,
			&column.ColumnDefault, &column.IsNullable, &column.DataType, &column.CharacterMaximumLength, &column.CharacterOctetLength,
			&column.NumericPrecision, &column.NumericScale, &column.CollationName, &column.ColumnType, &column.ColumnKey, &column.ColumnKey,
			&column.Extra, &column.Privileges, &column.ColumnComment)
		if err != nil {
			log.Fatal(err)
		}

		columns = append(columns, column)
	}

	return columns
}

func GenerateStructs(columns []*Column) { //逆向工程所有的表
	tableName := ""

	var tableColumns []*Column
	for _, v := range columns {
		if v.TableName.String != tableName { //新的一个表
			if tableName != "" {
				GenerateStruct(tableColumns)
			}
			tableName = v.TableName.String
			tableColumns = make([]*Column, 0)
		}
		tableColumns = append(tableColumns, v)
	}
}
func GenerateStruct(columns []*Column) string { //逆向工程一个表
	structName := GetStructName(columns[0].TableName.String)

	structContent := fmt.Sprintf("type %s struct{\n", structName)

	structContent += "}\n"

	fmt.Println(structContent)
	return structContent
}

func GetStructName(tableName string) string { //表名转换到类名
	tableName = strings.Replace(tableName, "t_", "", 1)
	names := strings.Split(tableName, "_")

	structName := ""
	for _, v := range names {
		structName = fmt.Sprintf("%s%s", structName, strings.Title(v))
	}

	return structName
}

func GetField(column Column) string {
	fieldContent := ""

	fieldName := GetFieldName(column.ColumnName.String)
	fieldContent += fieldName

	return fieldContent
}
func GetFieldName(columnName string) string {
	names := strings.Split(columnName, "_")
	fieldName := ""
	for _, v := range names {
		fieldName = fmt.Sprintf("%s%s", fieldName, strings.Title(v))
	}

	return fieldName
}

func main() {
	//读取config.json
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	//读取配置文件
	LoadConfig(data)

	//连接数据库，查询columns表
	columns := GetColumns()

	//开始生成struct
	GenerateStructs(columns)
}
