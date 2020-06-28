package main

import (
	"database/sql"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-oci8"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
)

const cfgPath string = "./config/"

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
	Get string `toml:"getInfo"`
}

type chargActInfoPreHyb struct {
	phNumb string
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

func getDbData(dsn string, querySql string, m map[string]string) {
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		panic(err)
	}
	if db == nil {
		fmt.Println("db is nil")
	}
	defer db.Close()

	rows, err := db.Query(querySql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var cai chargActInfoPreHyb
		rows.Scan(&cai.phNumb)
		m[cai.phNumb] = cai.phNumb
	}
}

func ctSoapPreHyb(phNumb string) string {
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
	soap += "            <AddBalance>-1</AddBalance>\n"
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
		fmt.Println(err)
	}
}

func main() {
	//	generateExcel(dbUser)

	var cfg cfgInfo
	getODbInfo(os.Args[1], &cfg)
	dsn := fmt.Sprintf("%s/%s@%s:%d/%s", cfg.Db.User, cfg.Db.Pwd, cfg.Db.IP, cfg.Db.Port, cfg.Db.Sid)
	fmt.Println(dsn)
	fmt.Println(cfg.WSURL)

	actInfo := make(map[string]string)
	getDbData(dsn, cfg.SrcSqlPreHyb.Get, actInfo)

	wg := &sync.WaitGroup{}
	limiter := make(chan bool, 20)
	for k, v := range actInfo {
		wg.Add(1)
		soapWS := ctSoapPreHyb(v)
		phNum := k
		limiter <- true
		go func() {
			defer wg.Done()
			res, err := http.Post(cfg.WSURL, "text/xml; charset=UTF-8", strings.NewReader(soapWS))
			if err != nil {
				//fmt.Printf("[error] act: %s, http post err: %s\n", phNum, err)
				fmt.Printf("[error] phNum: %s, http post err: %s\n", phNum, err)
				runtime.Goexit()
			}

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				//fmt.Printf("[error] act: %s, ioutil readAll err: %s\n", phNum, err)
				fmt.Printf("[error] phNum: %s, ioutil readAll err: %s\n", phNum, err)
				runtime.Goexit()
			}

			if res.StatusCode != http.StatusOK {
				//fmt.Printf("[error] act: %s, webService request failed, status is: %d\n", phNum, res.StatusCode)
				fmt.Printf("[error] phNum: %s, webService request failed, status is: %d. >>Response body is: %s",
					phNum, res.StatusCode, string(data))
				runtime.Goexit()
			}

			//fmt.Printf("[succeed] act: %s, webService response: %s\n", phNum, string(data))
			fmt.Printf("[succeed] phNum: %s, webService response: %s\n", phNum, string(data))

			res.Body.Close()
			<-limiter
			runtime.Goexit()
		}()
	}

	wg.Wait()

	fmt.Println("all finished.")
}
