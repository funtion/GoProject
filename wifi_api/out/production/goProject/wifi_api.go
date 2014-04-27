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
	"crypto/sha512"
	"hash"
	"crypto/sha256"
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
	PassHash []byte
}

var (
	hash map[string]CacheData
)

func cleanCache(){
	for {
		time.Sleep(5*60*time.Second)
		go func() {
			now := time.Now()
			fmt.Println("staarting dc",now)
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
    res,err:=client.PostForm(LOGIN_URL,data)
	_,failed :=res.Header["Content-Length"]
	if err!=nil || failed {
		m = Res{"false","loginfailed",""}
	}else {
		res2,err :=client.Get(INFO_URL)

		if err!=nil 	{
			m = Res{"false","getInfofailed",""}
		} else {
			body,_:=ioutil.ReadAll(res2.Body)
			bodyStr:=string(body)
			splited_body:=strings.Split(bodyStr,"\t")
			realData:=strings.Trim(splited_body[255],"\n")//255 is a magic number
			m = Res{"true","",realData}
		}

	}

	ans,_:= json.Marshal(m)
	chanel <- string(ans)
}
func log(username,res string,duration time.Duration ){
	durstr :=duration.String()
	logStr := time.Now().String()+": "+username+" "+res+" "+durstr+"\n"
	fmt.Print(logStr)
	file,err:=os.OpenFile("log.txt", os.O_RDWR|os.O_APPEND,0660);
	defer file.Close()
	if err==nil{
		file.WriteString(logStr)
	}else if os.IsNotExist(err) {
		file,err = os.Create("log.txt")
		if err==nil {
			file.WriteString(logStr)
		}else{
			fmt.Println("create log file failed")
		}
	}else{
		fmt.Println("open log file error!")
	}

}
func dataHandler(w http.ResponseWriter, r *http.Request){
	startTime := time.Now()
	username,password:=r.FormValue("username"),r.FormValue("password")
	passHash :=sha512.Sum512([]byte(password))
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
	endTime := time.Now()
	fmt.Fprintf(w,res)
	log(username,res,endTime.Sub(startTime))
}
func main() {
	//TODO log
	hash = make(map[string]CacheData)
	http.HandleFunc("/",dataHandler)
	http.ListenAndServe(":8888",nil)
	cleanCache()
}
