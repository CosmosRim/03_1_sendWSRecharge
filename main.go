package main

import (
	"database/sql"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-oci8"
	"os"
)

const cfgPath string = "./config/"

type cfgInfo struct {
	Db     oDbInfo `toml:"oDbInfo"`
	SrcSql srcSql
}

type oDbInfo struct {
	User string `toml:"oDbUser"`
	Pwd  string `toml:"oDbPwd"`
	IP   string `toml:"oDbIp"`
	Port int    `toml:"oDbPort"`
	Sid  string `toml:"oDbSid"`
}

type srcSql struct {
	Get string `toml:"getInfo"`
}

type chargActInfo struct {
	RUT     string
	ActCode string
}

func getODbInfo(fl string, st *cfgInfo) {
	if _, err := toml.DecodeFile(cfgPath+fl, st); err != nil {
		panic(err)
	}
}

func getDbData(dsn string, querySql string, m map[string]string) {
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query(querySql)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var cai chargActInfo
		rows.Scan(&cai.RUT, &cai.ActCode)
		m[cai.ActCode] = cai.RUT
	}
}

func generateExcel(databaseUser string) {
	f := excelize.NewFile()
	index := f.NewSheet("Sheet2")

	f.SetCellValue("Sheet1", "A2", "初次见面，Excel module。")
	f.SetCellValue("Sheet2", "B2", databaseUser)

	f.SetActiveSheet(index)

	if err := f.SaveAs("first.xlsx"); err != nil {
		fmt.Println(err)
	}
}

func main() {
	//	generateExcel(dbUser)

	var cfg cfgInfo
	getODbInfo(os.Args[1], &cfg)
	dsn := fmt.Sprintf("%s/%s@%s:%d/%s", cfg.Db.User, cfg.Db.Pwd, cfg.Db.IP, cfg.Db.Port, cfg.Db.Sid)
	fmt.Println(dsn)

	actInfo := make(map[string]string)
	getDbData(dsn, cfg.SrcSql.Get, actInfo)
	for k, v := range actInfo {
		fmt.Println(k, v)
	}
}
