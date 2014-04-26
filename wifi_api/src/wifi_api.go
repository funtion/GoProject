package main

import (
	"fmt"
	"net/url"
	"net/http"
	"net/http/cookiejar"
	"io/ioutil"
	"strings"
	"encoding/json"
	"time"
	"os"
)
const (
	LOGIN_URL = "https://nic.seu.edu.cn/selfservice/campus_login.php"
	INFO_URL = "https://nic.seu.edu.cn/selfservice/service_manage_status_web.php"
)
type Res struct {
	Success string
	Reason string
	Data string
}

type CacheData struct {
	LastTime time.Time
	Data string
}

var (
	hash map[string]CacheData
)

func cleanCache(){
	for {
		time.Sleep(5*60*time.Second)
		go func() {
			now := time.Now()
			for key,value := range hash {
				if now.Sub(value.LastTime) > 5*60*time.Second {
					delete(hash,key)
				}
			}
		}()
	}
}

func post(username,password string,chanel chan string) {
	data := make(url.Values)
	data.Set("username",username)
	data.Set("password",password)
	cookieJar,_ :=cookiejar.New(nil)
	client :=&http.Client{
		Jar: cookieJar,
	}
	var m Res
    _,err:=client.PostForm(LOGIN_URL,data)
	if err!=nil{
		m = Res{"false","loginfailed",""}
	} else {
		res2,err :=client.Get(INFO_URL)
		if err!=nil {
			m = Res{"false","loginfailed",""}
		} else {
			body,_:=ioutil.ReadAll(res2.Body)
			bodyStr:=string(body)
			splited_body:=strings.Split(bodyStr,"\t")
			realData:=strings.Trim(splited_body[255],"\n")//255 is magic number
			m = Res{"true","",realData}
		}

	}

	ans,_:= json.Marshal(m)
	chanel <- string(ans)
}
func log(username,res string){
	logStr := time.Now().String()+": "+username+" "+res
	file,err:=os.OpenFile("D:\\log.txt", os.O_RDWR|os.O_APPEND,0660);
	if err!=nil{
		file.WriteString(logStr)
	}
	file.Close()
}
func dataHandler(w http.ResponseWriter, r *http.Request){
	username,password:=r.FormValue("username"),r.FormValue("password")
	var res string;
	_,ok:=hash[username]
	if !ok  || (time.Now().Sub(hash[username].LastTime ) > time.Second*60*3) {
		chanel := make(chan string)
		go post(username,password,chanel)
		res = <-chanel
		go func(){
			hash[username] = CacheData {LastTime:time.Now(),Data:res}
		}()

	} else {
		res = hash[username].Data
	}
	fmt.Fprintf(w,res)
	log(username,res)
}
func main() {
	//TODO log
	//TODO cache
	hash = make(map[string]CacheData)
	http.HandleFunc("/",dataHandler)
	http.ListenAndServe(":8888",nil)
	cleanCache()
}
