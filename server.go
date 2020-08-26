package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/gorilla/mux"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func getCrv(w http.ResponseWriter, r *http.Request) {
	var outStr string
	inBody, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	doc := etree.NewDocument()
	_ = doc.ReadFromString(string(inBody))
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
	middleName := requestElm.CreateElement("MiddleName")
	middleName.CreateText(root.SelectElement("MiddleName").Text())
	passport := requestElm.CreateElement("Passport")
	passport.CreateText(root.SelectElement("Passport").Text())
	phone := requestElm.CreateElement("Phone")
	phone.CreateText(root.SelectElement("Phone").Text())
	responseElm := rootOut.CreateElement("Response")
	dateTime := responseElm.CreateElement("DateTime")
	dateTime.CreateText(getDate())
	result := responseElm.CreateElement("Result")
	result.CreateText("N")

	w.Header().Set("Content-Type", "application/xml;charset=windows-1251")
	outStr, _ = docOut.WriteToString()
	outResp, _ := toWin1251(outStr)
	fmt.Fprint(w, outResp)
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

func getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}

func getVersion(w http.ResponseWriter, r *http.Request, branch string) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"data": {"version": "`+branch+`"}, "error": {}}`)
}

func main() {
	listenPort := "8080"
	if len(os.Getenv("GO_OTHER_MOCKS_PORT")) > 0 {
		listenPort = os.Getenv("GO_OTHER_MOCKS_PORT")
	}
	var branch string
	if len(os.Getenv("GO_OTHER_MOCK_BRANCH")) > 0 {
		branch = os.Getenv("GO_OTHER_MOCK_BRANCH")
	}
	log.Print("App start on port ", listenPort)
	r := mux.NewRouter()
	r.HandleFunc("/crv", func(w http.ResponseWriter, r *http.Request) { getCrv(w, r) }).Methods("POST")
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) { getStatus(w, r) }).Methods("GET")
	r.HandleFunc("/api/system/version", func(w http.ResponseWriter, r *http.Request) { getVersion(w, r, branch) }).Methods("GET")
	log.Fatal(http.ListenAndServe(":"+listenPort, r))
}
