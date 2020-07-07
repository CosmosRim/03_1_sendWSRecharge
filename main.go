package main

import (
	"database/sql"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-oci8"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

const cfgPath string = "./config/"
const logPath string = "./log/"

type cfgInfo struct {
	WSURL        string
	Db           oDbInfo `toml:"oDbInfo"`
	SrcSqlPreHyb srcSqlPreHyb
	SrcSqlPost   srcSqlPost
}
type oDbInfo struct {
	User string `toml:"oDbUser"`
	Pwd  string `toml:"oDbPwd"`
	IP   string `toml:"oDbIp"`
	Port int    `toml:"oDbPort"`
	Sid  string `toml:"oDbSid"`
}
type srcSqlPost struct {
	Get string `toml:"getInfo"`
}
type srcSqlPreHyb struct {
	GetNum  string
	GetInfo string
}

type chargActInfoPreHyb struct {
	phNumb string
	chgAmt string
}
type chargActInfoPost struct {
	transSN string
	RUT     string
	ActCode string
	Folio   string
}

func getODbInfo(fl string, st *cfgInfo) {
	if _, err := toml.DecodeFile(cfgPath+fl, st); err != nil {
		panic(err)
	}
}

func getDbNum(dsn string, querySql string) uint32 {
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		panic(err)
	}
	if db == nil {
		log.Println("db is nil")
	}
	defer db.Close()

	rows, err := db.Query(querySql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var i uint32
	for rows.Next() {
		rows.Scan(&i)
	}
	return i
}
func getDbData(dsn string, querySql string, s []chargActInfoPreHyb) {
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		panic(err)
	}
	if db == nil {
		log.Println("db is nil")
	}
	defer db.Close()

	rows, err := db.Query(querySql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var i uint32
	for rows.Next() {
		var caip chargActInfoPreHyb
		rows.Scan(&caip.phNumb, &caip.chgAmt)
		s[i] = caip
		i++
	}
}

func addErrAct(u *uint32, s []chargActInfoPreHyb, info chargActInfoPreHyb) {
	s[*u] = info
	atomic.AddUint32(u, 1)
}

func ctSoapPreHyb(phNumb string, chgAmt string) string {
	soap := "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n"
	soap += "<soapenv:Envelope xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:xsd=\"http://com.ztesoft.zsmart/xsd\">\n"
	soap += "    <soapenv:Header>\n"
	soap += "        <xsd:AuthHeader>\n"
	soap += "            <Username>NY</Username>\n"
	soap += "            <Password>smart</Password>\n"
	soap += "            <ChannelCode>Devetel</ChannelCode>\n"
	soap += "        </xsd:AuthHeader>\n"
	soap += "    </soapenv:Header>\n"
	soap += "    <soapenv:Body>\n"
	soap += "        <xsd:RechargeBO>\n"
	soap += "            <TransactionSN>1589288094591016</TransactionSN>\n"
	soap += "            <DealerCode>JJD</DealerCode>\n"
	soap += "            <PaymentChannelID>23</PaymentChannelID>\n"
	soap += "            <ServiceNumber>" + phNumb + "</ServiceNumber>\n"
	soap += "            <AccountCode/>\n"
	soap += "            <AcctResCode/>\n"
	soap += "            <AddBalance>" + chgAmt + "</AddBalance>\n"
	soap += "            <AddDays/>\n"
	soap += "            <OperationStaff/>\n"
	soap += "            <InactiveDuration/>\n"
	soap += "        </xsd:RechargeBO>\n"
	soap += "    </soapenv:Body>\n"
	soap += "</soapenv:Envelope>\n"
	return soap
}

func ctSoapPost(transSN string, rut string, actCode string, folio string) string {
	soap := "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n"
	soap += "<soapenv:Envelope xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:xsd=\"http://com.ztesoft.zsmart/xsd\">\n"
	soap += "    <soapenv:Header>\n"
	soap += "        <xsd:AuthHeader>\n"
	soap += "            <Username>NY</Username>\n"
	soap += "            <Password>smart</Password>\n"
	soap += "            <!--Optional:-->\n"
	soap += "            <ChannelCode>Devetel</ChannelCode>\n"
	soap += "        </xsd:AuthHeader>\n"
	soap += "    </soapenv:Header>\n"
	soap += "    <soapenv:Body>\n"
	soap += "        <xsd:PaymentRequest><TransactionSN>" + transSN + "</TransactionSN>\n"
	soap += "            <!--Optional:-->\n"
	soap += "            <TransactionDesc xsi:nil=\"true\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"/>\n"
	soap += "            <RUT>" + rut + "</RUT>\n"
	soap += "            <AccountCode>" + actCode + "</AccountCode>\n"
	soap += "            <!--Optional:-->\n"
	soap += "            <ServiceNumber xsi:nil=\"true\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"/>\n"
	soap += "            <PaymentType>1</PaymentType>\n"
	soap += "            <!--Optional:-->\n"
	soap += "            <FolioID>" + folio + "</FolioID>\n"
	soap += "            <PaymentAmount>12580</PaymentAmount>\n"
	soap += "            <CURRENCY>CLP</CURRENCY>\n"
	soap += "            <PostDate>2020-06-18T12:34:56</PostDate>\n"
	soap += "            <PaymentChannelID>202-11</PaymentChannelID>\n"
	soap += "            <PaymentMethod>99</PaymentMethod>\n"
	soap += "            <PaymentGatewayChannel>DEVETEL</PaymentGatewayChannel>\n"
	soap += "            <!--Optional:-->\n"
	soap += "            <CreditCardType xsi:nil=\"true\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"/>\n"
	soap += "            <!--Optional:-->\n"
	soap += "            <BankCode xsi:nil=\"true\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"/>\n"
	soap += "        </xsd:PaymentRequest>\n"
	soap += "    </soapenv:Body>\n"
	soap += "</soapenv:Envelope>\n"
	return soap
}

func generateExcel(databaseUser string) {
	f := excelize.NewFile()
	index := f.NewSheet("Sheet2")

	f.SetCellValue("Sheet1", "A2", "初次见面，Excel module。")
	f.SetCellValue("Sheet2", "B2", databaseUser)

	f.SetActiveSheet(index)

	if err := f.SaveAs("first.xlsx"); err != nil {
		log.Println(err)
	}
}

func main() {
	//	generateExcel(dbUser)

	//remove log folder and
	if _, err := os.Stat(logPath); err == nil {
		log.Println("Remove exists old folder : " + logPath)
		os.RemoveAll(logPath)
	}

	if err := os.Mkdir(logPath, os.ModePerm); err != nil {
		log.Println("Creaete log folder: " + logPath + " failed.")
	} else {
		log.Println("Creaete log folder: " + logPath + " succeed.")
	}

	//craete log file and start a log object
	logFileName := logPath + "WS.log"
	if _, err := os.Stat(logFileName); err == nil {
		log.Println("Remove exists old log file: " + logFileName)
		os.Remove(logFileName)
	}

	logFile, err := os.Create(logFileName)
	defer logFile.Close()
	if err != nil {
		log.Fatalf("Create log file with error: %s", err)
	}

	logOb := log.New(logFile, "", log.LstdFlags)

	var cfg cfgInfo
	getODbInfo(os.Args[1], &cfg)
	dsn := fmt.Sprintf("%s/%s@%s:%d/%s", cfg.Db.User, cfg.Db.Pwd, cfg.Db.IP, cfg.Db.Port, cfg.Db.Sid)
	logOb.Println("warming engine...")
	logOb.Println("dsn is: " + dsn)
	logOb.Println("webService URL is: " + cfg.WSURL)

	var sliNum uint32
	sliNum = getDbNum(dsn, cfg.SrcSqlPreHyb.GetNum)
	logOb.Printf("Amount of recharge records: %d\n", sliNum)

	actInfo := make([]chargActInfoPreHyb, sliNum)
	getDbData(dsn, cfg.SrcSqlPreHyb.GetInfo, actInfo)
	logOb.Println("Got src data, start package and post.")
	log.Println("Got src data, start package and post.")

	errActInfo := make([]chargActInfoPreHyb, sliNum)
	var errCount uint32

	wg := &sync.WaitGroup{}
	muErrActInfo := &sync.Mutex{}
	limiter := make(chan bool, 26)
	for _, v := range actInfo {
		wg.Add(1)
		value := v
		soapWS := ctSoapPreHyb(value.phNumb, value.chgAmt)
		limiter <- true
		go func() {
			defer wg.Done()
			defer func() { <-limiter }()
			res, err := http.Post(cfg.WSURL, "text/xml; charset=UTF-8", strings.NewReader(soapWS))
			defer res.Body.Close()
			if err != nil {
				muErrActInfo.Lock()
				addErrAct(&errCount, errActInfo, value)
				muErrActInfo.Unlock()
				logOb.Printf("[error] phNum: %s , amt: %s , http post err: %s\n", value.phNumb, value.chgAmt, err)
				runtime.Goexit()
			}

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				muErrActInfo.Lock()
				addErrAct(&errCount, errActInfo, value)
				muErrActInfo.Unlock()
				logOb.Printf("[error] phNum: %s , amt: %s , ioutil readAll err: %s\n", value.phNumb, value.chgAmt, err)
				runtime.Goexit()
			}

			if res.StatusCode != http.StatusOK {
				muErrActInfo.Lock()
				addErrAct(&errCount, errActInfo, value)
				muErrActInfo.Unlock()
				logOb.Printf("[error] phNum: %s , amt: %s , webService request failed, status is: %d. >>Response body is: %s",
					value.phNumb, value.chgAmt, res.StatusCode, string(data))
				runtime.Goexit()
			}

			logOb.Printf("[succeed] phNum: %s , amt: %s , webService response: %s\n", value.phNumb, value.chgAmt, string(data))

			runtime.Goexit()
		}()
	}

	wg.Wait()

	logOb.Println("first batch post finished.")
	logFile.Close()

	//start to repost untill no more error bak. each batch have their own log.
	logBatch := 1
	for errCount > 0 {
		errRange := errCount
		errCount = 0

		//create log file for each repost
		logReFileName := fmt.Sprintf("%sWSRe%d.log", logPath, logBatch)
		if _, err := os.Stat(logReFileName); err == nil {
			log.Println("Remove exists old log file: " + logReFileName)
			os.Remove(logReFileName)
		}

		logReFile, err := os.Create(logReFileName)
		if err != nil {
			log.Fatalf("Create log file with error: %s", err)
		}

		logObRe := log.New(logReFile, "", log.LstdFlags)

		//repeat post till no more error bak
		var i uint32
		for i < errRange {
			errActInfoI := errActInfo[i]
			//if errActInfoI.phNumb == "" {
			//	i++
			//	continue
			//}
			soapWS := ctSoapPreHyb(errActInfoI.phNumb, errActInfoI.chgAmt)

			wg.Add(1)
			limiter <- true

			go func() {
				defer wg.Done()
				defer func() { <-limiter }()
				res, err := http.Post(cfg.WSURL, "text/xml; charset=UTF-8", strings.NewReader(soapWS))
				defer res.Body.Close()
				if err != nil {
					muErrActInfo.Lock()
					addErrAct(&errCount, errActInfo, errActInfoI)
					muErrActInfo.Unlock()
					logObRe.Printf("[error] phNum: %s , amt: %s , http post err: %s\n", errActInfoI.phNumb, errActInfoI.chgAmt, err)
					runtime.Goexit()
				}

				data, err := ioutil.ReadAll(res.Body)
				if err != nil {
					muErrActInfo.Lock()
					addErrAct(&errCount, errActInfo, errActInfoI)
					muErrActInfo.Unlock()
					logObRe.Printf("[error] phNum: %s , amt: %s , ioutil readAll err: %s\n", errActInfoI.phNumb, errActInfoI.chgAmt, err)
					runtime.Goexit()
				}

				if res.StatusCode != http.StatusOK {
					muErrActInfo.Lock()
					addErrAct(&errCount, errActInfo, errActInfoI)
					muErrActInfo.Unlock()
					logObRe.Printf("[error] phNum: %s , amt: %s , webService request failed, status is: %d. >>Response body is: %s\n",
						errActInfoI.phNumb, errActInfoI.chgAmt, res.StatusCode, string(data))
					runtime.Goexit()
				}

				logObRe.Printf("[succeed] phNum: %s , amt: %s , webService response: %s\n", errActInfoI.phNumb, errActInfoI.chgAmt, string(data))
				runtime.Goexit()
			}()

			i++
		}
		wg.Wait()
		logObRe.Printf("WSRe%d.log finished", logBatch)
		logReFile.Close()
		logBatch++
	}

	log.Println("all finished.")
}
