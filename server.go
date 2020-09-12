package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/fasthttp/router"
	pudge "github.com/recoilme/pudge"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var (
	branch string
	db     *pudge.Db
)

func getCrv(ctx *fasthttp.RequestCtx) {
	var outStr string
	strBody, _ := toUtf8(string(ctx.Request.Body()))
	log.Print(strings.Replace(strBody, "\n", "", -1))

	doc := etree.NewDocument()
	_ = doc.ReadFromString(strBody)
	root := doc.SelectElement("CRV")

	docOut := etree.NewDocument()
	docOut.CreateProcInst("xml", `version="1.0" encoding="windows-1251"`)
	rootOut := docOut.CreateElement("CRV")
	requestElm := rootOut.CreateElement("Request")
	login := requestElm.CreateElement("Login")
	login.CreateText(root.SelectElement("Login").Text())
	password := requestElm.CreateElement("Password")
	password.CreateText(root.SelectElement("Password").Text())
	lastName := requestElm.CreateElement("LastName")
	lastName.CreateText(root.SelectElement("LastName").Text())
	firstName := requestElm.CreateElement("FirstName")
	firstName.CreateText(root.SelectElement("FirstName").Text())
	if inMiddleName := root.SelectElement("MiddleName"); inMiddleName != nil {
		middleName := requestElm.CreateElement("MiddleName")
		middleName.CreateText(root.SelectElement("MiddleName").Text())
	}
	passport := requestElm.CreateElement("Passport")
	passport.CreateText(root.SelectElement("Passport").Text())
	phone := requestElm.CreateElement("Phone")
	phone.CreateText(root.SelectElement("Phone").Text())
	responseElm := rootOut.CreateElement("Response")
	dateTime := responseElm.CreateElement("DateTime")
	dateTime.CreateText(getDate())
	result := responseElm.CreateElement("Result")
	result.CreateText("N")

	ctx.Response.Header.Set("Content-Type", "application/xml;charset=windows-1251")
	outStr, _ = docOut.WriteToString()
	outResp, _ := toWin1251(outStr)
	fmt.Fprint(ctx, outResp)
}

func getDate() string {
	tm := time.Now()
	return tm.Format("02.01.2006 15:04:05")
}

func toUtf8(text string) (s string, err error) {
	sr := strings.NewReader(text)
	tr := transform.NewReader(sr, charmap.Windows1251.NewDecoder())
	buf, err := ioutil.ReadAll(tr)
	if err != err {
		return
	}
	s = string(buf)
	return
}

func toWin1251(text string) (s string, err error) {
	sr := strings.NewReader(text)
	tr := transform.NewReader(sr, charmap.Windows1251.NewEncoder())
	buf, err := ioutil.ReadAll(tr)
	if err != err {
		return
	}
	s = string(buf)
	return
}

func getFastCheck(ctx *fasthttp.RequestCtx) {
	var filepath string
	log.Print(string(ctx.RequestURI()))
	switch string(ctx.RequestURI()) {
	case "/ws/service?wsdl":
		filepath = "./fastcheck.wsdl"
	case "/ws/service?xsd":
		filepath = "./fastcheck.xsd"
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Print("Read file ", filepath, " ERROR ", err)
		ctx.Error("Read file error", fasthttp.StatusInternalServerError)
	}
	ctx.Response.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	fmt.Fprint(ctx, string(file))
}

func postFastCheck(ctx *fasthttp.RequestCtx) {
	var s string
	doc := etree.NewDocument()
	err := doc.ReadFromBytes(ctx.Request.Body())
	if err != nil {
		log.Print(err)
		ctx.Error("Read request xml error", fasthttp.StatusBadRequest)
	}
	params := doc.FindElements("//soap-env:Envelope/soap-env:Body/fas:fasRequest")[0]
	birthDate := params.SelectElement("fas:dob").Text()
	passNumber := params.SelectElement("fas:passport").Text()

	out, err := readCache(passNumber)
	if err != nil {

		template := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><soapenv:Envelope xmlns:fas=\"http://mbtc.ru/fas\" xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\"><soapenv:Body><fas:fasResponse><fas:date>{{curDate}}</fas:date><fas:loanRating><fas:subjectBirthDate>{{birthDate}}</fas:subjectBirthDate><fas:subjectBirthDate>{{birthDate}}</fas:subjectBirthDate><fas:count>{{count}}</fas:count><fas:guarantee>0</fas:guarantee><fas:limit>{{limit}}</fas:limit><fas:balance>{{balance}}</fas:balance><fas:maxCloseDt>{{maxDate}}</fas:maxCloseDt><fas:delay>{{delay}}</fas:delay><fas:totalOverdue>{{totalOverdue}}</fas:totalOverdue><fas:maxDeleay>{{maxDeleay}}</fas:maxDeleay><fas:closedNegative>{{closedNegative}}</fas:closedNegative><fas:countDue30_60InOpenedAccs>{{countDue30}}</fas:countDue30_60InOpenedAccs><fas:countDue30_60InClosedAccs>0</fas:countDue30_60InClosedAccs><fas:countDue60_90InOpenedAccs>0</fas:countDue60_90InOpenedAccs><fas:countDue60_90InClosedAccs>0</fas:countDue60_90InClosedAccs><fas:countDue90PlusInOpenedAccs>{{countDue90o}}</fas:countDue90PlusInOpenedAccs><fas:countDue90PlusInClosedAccs>{{countDue90c}}</fas:countDue90PlusInClosedAccs></fas:loanRating><fas:docCheck><fas:found>Y</fas:found><fas:hasNewer>N</fas:hasNewer><fas:lost>N</fas:lost><fas:invalid>N</fas:invalid><fas:wanted>N</fas:wanted></fas:docCheck><fas:inquiries><fas:week>{{week}}</fas:week><fas:twoWeeks>{{twoWeeks}}</fas:twoWeeks><fas:month>{{month}}</fas:month><fas:overall>{{overall}}</fas:overall></fas:inquiries><fas:checks><fas:check><fas:date>2020-09-12T13:07:37</fas:date><fas:isOwn>Y</fas:isOwn><fas:loanType>Другой</fas:loanType><fas:loanAmount>100000</fas:loanAmount><fas:loanDuration>более 60 дней</fas:loanDuration></fas:check></fas:checks></fas:fasResponse></soapenv:Body></soapenv:Envelope>"
		t := fasttemplate.New(template, "{{", "}}")
		s = t.ExecuteString(map[string]interface{}{
			"curDate":        time.Now().Format("2006-01-02"),
			"birthDate":      birthDate,
			"count":          strconv.Itoa(randInt(1, 5)),
			"limit":          strconv.Itoa(randInt(1000, 300000)),
			"balance":        strconv.Itoa(randInt(100, 300000)),
			"maxDate":        strconv.Itoa(randInt(2030, 2100)) + "-01-01",
			"delay":          strconv.Itoa(randInt(1, 5)),
			"totalOverdue":   strconv.Itoa(randInt(0, 100000)),
			"maxDeleay":      strconv.Itoa(randInt(1, 5)),
			"closedNegative": "Y",
			"countDue30":     strconv.Itoa(randInt(0, 1)),
			"countDue90o":    strconv.Itoa(randInt(0, 30)),
			"countDue90c":    strconv.Itoa(randInt(0, 30)),
			"week":           strconv.Itoa(randInt(0, 3)),
			"twoWeeks":       strconv.Itoa(randInt(0, 6)),
			"month":          strconv.Itoa(randInt(1, 13)),
			"overall":        strconv.Itoa(randInt(1, 125)),
		})
		_ = writeCache(passNumber, s)
	} else {
		s = out
	}
	ctx.Response.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	fmt.Fprint(ctx, s)
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func readCache(param string) (string, error) {
	var err error
	var data string
	db.Get(param, &data)
	if len(data) == 0 {
		err = errors.New("Error, unable to read value from cache")
	}
	return data, err
}

func writeCache(param string, data string) error {
	var err error
	var checkData string
	db.Set(param, data)
	db.Get(param, &checkData)
	if checkData != data {
		err = errors.New("Error, unable to add value to cache")
	}
	return err
}

func getStatus(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	fmt.Fprint(ctx, `{"status":"ok"}`)
}

func getVersion(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	fmt.Fprint(ctx, `{"data": {"version": "`+branch+`"}, "error": {}}`)
}

func main() {
	var err error
	listenPort := "8080"
	if len(os.Getenv("GO_OTHER_MOCKS_PORT")) > 0 {
		listenPort = os.Getenv("GO_OTHER_MOCKS_PORT")
	}
	if len(os.Getenv("GO_OTHER_MOCKS_BRANCH")) > 0 {
		branch = os.Getenv("GO_OTHER_MOCKS_BRANCH")
	}
	cfg := &pudge.Config{StoreMode: 2}
	db, err = pudge.Open("", cfg)
	if err != nil {
		log.Panic(err)
	}
	router := router.New()
	router.POST("/crv", getCrv)
	router.POST("/ws/service", postFastCheck)
	router.GET("/ws/service", getFastCheck)
	router.GET("/status", getStatus)
	router.GET("/api/system/version", getVersion)
	server := &fasthttp.Server{
		Handler: router.Handler,
		//		MaxRequestBodySize: 100 << 20,
		Concurrency:      1024 * 30,
		MaxConnsPerIP:    1024 * 10,
		DisableKeepalive: true,
	}
	log.Print("App start on port ", listenPort)
	log.Fatal(server.ListenAndServe(":" + listenPort))
}
