package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var branch string

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
		log.Print("Read file ", filepath, " ERROR", err)
		ctx.Error("Read file error", fasthttp.StatusInternalServerError)
	}
	ctx.Response.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	fmt.Fprint(ctx, string(file))
}

func postFastCheck(ctx *fasthttp.RequestCtx) {
	doc := etree.NewDocument()
	err := doc.ReadFromBytes(ctx.Request.Body())
	if err != nil {
		log.Print(err)
		ctx.Error("Read request xml error", fasthttp.StatusBadRequest)
	}
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
	listenPort := "8080"
	if len(os.Getenv("GO_OTHER_MOCKS_PORT")) > 0 {
		listenPort = os.Getenv("GO_OTHER_MOCKS_PORT")
	}
	if len(os.Getenv("GO_OTHER_MOCKS_BRANCH")) > 0 {
		branch = os.Getenv("GO_OTHER_MOCKS_BRANCH")
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
