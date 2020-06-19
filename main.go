package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-oci8"
)

const cfgPath string = "./config/"

type cfgInfo struct {
	Db oDbInfo `toml:"oDbInfo"`
}

type oDbInfo struct {
	User string `toml:"oDbUser"`
	Pwd  string `toml:"oDbPwd"`
	IP   string `toml:"oDbIp"`
	Port int    `toml:"oDbPort"`
	Sid  string `toml:"oDbSid"`
}

func getODbInfo(fl string, st *cfgInfo) {
	if _, err := toml.DecodeFile(cfgPath+fl, st); err != nil {
		panic(err)
	}
}

func getDbData() string {
	if len(os.Args) != 2 {
		fmt.Printf("ERROR: Please provide a DSN string in ONE argument:\n\n")
		fmt.Println("Shell-Conversion into DSN string:")
		fmt.Println("  sqlplus sys/password@tnsentry as sysdba   =>   sys/password@tnsentry?as=sysdba")
		fmt.Println("  sqlplus / as sysdba                       =>   sys/.@?as=sysdba")
		fmt.Println("instead of the tnsentry, you can also use the hostname of the IP.")
		os.Exit(1)
	}
	os.Setenv("NLS_LANG", "")

	db, err := sql.Open("oci8", os.Args[1])
	if err != nil {
		fmt.Println(err)
		return "connet err"
	}
	defer db.Close()
	//fmt.Println()
	var user string
	err = db.QueryRow("select user from dual").Scan(&user)
	if err != nil {
		fmt.Println(err)
		return "query err"
	}
	fmt.Printf("Connect successfully 'as sysdba'. Current user is: %v\n\n", user)
	return user
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
	//	dbUser := getDbData()
	//	generateExcel(dbUser)

	var usage cfgInfo
	getODbInfo(os.Args[1], &usage)
	fmt.Printf("dsn is : %s/%s@%s:%d/%s", usage.Db.User, usage.Db.Pwd, usage.Db.IP, usage.Db.Port, usage.Db.Sid)
}
