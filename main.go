package main

import (
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	Text     = "text"
	Location = "location"
	Image    = "image"
	Link     = "link"
	Event    = "event"
	Music    = "music"
	News     = "news"
)

type Config struct {
	Micromsg struct {
		AppId          string
		AppSecret      string
		Token          string
		EncodingASEKey string
	}
	Http struct {
		Listenip string
		Port     string
	}
}

type msgBase struct {
	ToUserName   string
	FromUserName string
	CreateTime   time.Duration
	MsgType      string
	Content      string
}

type Request struct {
	XMLName                xml.Name `xml:"xml"`
	msgBase                         //base struct
	Location_x, Location_y float32
	Scale                  int
	Label                  string
	PicUrl                 string
	MsgId                  int
}

type Response struct {
	XMLName xml.Name `xml:"xml"`
	msgBase
	ArticleCount int     `xml:",omitempty"`
	Articles     []*item `xml:"Articles>item,omitempty"`
	FuncFlag     int
}

type item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string
	Description string
	PicUrl      string
	Url         string
}

func weixinEvent(w http.ResponseWriter, r *http.Request) {
	if stat := weixinCheckSignature(w, r); stat == false {
		log.Println("auth failed")
		return
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println(string(body))
	var wrep *Request
	if wrep, err = DecodeRequest(body); err != nil {
		log.Fatal(err)
		return
	}
	wresp, err := dealwith(wrep)
	if err != nil {
		log.Fatal(err)
		return
	}

	data, err := wresp.Encode()
	if err != nil {
		log.Printf("error:%v\n", err)
	}
	log.Println(string(data))

}

func DecodeRequest(data []byte) (req *Request, err error) {
	req = &Request{}
	if err = xml.Unmarshal(data, req); err != nil {
		return
	}
	req.CreateTime *= time.Second
	return

}

func str2sha1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))

}

func NewResponse() (resp *Response) {
	resp = &Response{}
	resp.CreateTime = time.Duration(time.Now().Unix())
	return
}

func (resp Response) Encode() (data []byte, err error) {
	resp.CreateTime = time.Duration(time.Now().Unix())
	data, err = xml.Marshal(resp)
	return

}
func dealwith(req *Request) (resp *Response, err error) {
	resp = &Response{}
	err = nil
	return

}

func weixinCheckSignature(w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	log.Println(r.Form)
	var signature string = strings.Join(r.Form["signature"], "")
	var timestamp string = strings.Join(r.Form["timestamp"], "")
	var nonce string = strings.Join(r.Form["nonce"], "")
	tmps := []string{conf.Micromsg.Token, timestamp, nonce}
	sort.Strings(tmps)
	tmpStr := tmps[0] + tmps[1] + tmps[2]
	tmp := str2sha1(tmpStr)
	if tmp == signature {
		return true
	}
	return false
}

func weixinAuth(w http.ResponseWriter, r *http.Request) {
	if weixinCheckSignature(w, r) == true {
		var echostr string = strings.Join(r.Form["echostr"], "")
		fmt.Fprintf(w, echostr)
	}
}

func weixinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Println("Get begin...")
		weixinAuth(w, r)
		fmt.Println("Get End...")

	}
}

var conf Config

func main() {
	r, err := os.Open("conf/config.json")
	if err != nil {
		log.Fatalln(err)
	}
	decoder := json.NewDecoder(r)

	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listening on port ", conf.Http.Port)

	http.HandleFunc("/check", weixinHandler)
	err = http.ListenAndServe(conf.Http.Listenip+":"+conf.Http.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

	log.Println(conf.Micromsg.AppId, conf.Micromsg.AppSecret, conf.Micromsg.Token, conf.Micromsg.EncodingASEKey)
}
