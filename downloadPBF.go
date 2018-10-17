/**
 * Created by bingqixuan on 2018/5/14.
 */
package main

import (
	"os"
	"fmt"
	"math"
	"log"
	"time"
	"net/http"
	"crypto/tls"
	"io"
	"encoding/xml"
	"io/ioutil"
	"sync"
)

type DownloadConfig struct {
	XMLName     xml.Name `xml:"servers"`
	Version     string   `xml:"version,attr"`
	MinX        int   `xml:"minX"`
	MinY        int   `xml:"minY"`
	MaxX        int   `xml:"maxX"`
	MaxY        int   `xml:"maxY"`
	MinLevel    int   `xml:"minLevel"`
	MaxLevel    int   `xml:"maxLevel"`
	CurrentLevel int  `xml:"currentLevel"`
	RootDir     string   `xml:"rootDir"`
	PreUrl      string   `xml:"preUrl"`
	MapboxToken string   `xml:"mapboxToken"`
}

func handleError(myerr interface{}){
	logfile, err := os.OpenFile("./error.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0)
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		os.Exit(-1)
	}
	defer logfile.Close()
	logger := log.New(logfile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)
	logger.Println(myerr)
}

func requestInfo(url string, result string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create New http Transport
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // disable verify
	}
	// Create Http Client
	client := &http.Client{Transport: transCfg}
	res, err := client.Get(url)
	if err != nil {
		handleError(err)
		panic(err)
	}

	f, err := os.Create(result)
	if err != nil {
		handleError(err)
		panic(err)
	}
	io.Copy(f, res.Body)
	res.Body.Close()
}

func main() {


	/**
	  配置参数
	 */
	configFile, err := os.Open("config/config.xml") // For read access.
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	defer configFile.Close()
	data, err := ioutil.ReadAll(configFile)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	v := DownloadConfig{}
	err = xml.Unmarshal(data, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	wg := sync.WaitGroup{}

	var minX = v.MinX         // 瓦片最小层级的最小行号，跟minLevel对应
	var minY = v.MinY          // 瓦片最小层级的最小列号，跟minLevel对应
	var maxX = v.MaxX         // 瓦片最小层级的最大行号，跟minLevel对应（不包括）
	var maxY = v.MaxY          // 瓦片最小层级的最大列号，跟minLevel对应（不包括）
	var minLevel = v.MinLevel       // 瓦片的最小层级
	var maxLevel = v.MaxLevel      // 瓦片的最大层级
	var currentLevel = v.CurrentLevel   // 可选择从哪一级开始下，此值介于minLevel和maxLevel之间
	var rootDir = v.RootDir   // 根目录文件夹
	var preUrl = v.PreUrl      // mapbox数据地址前缀
	var mapboxToken = v.MapboxToken   // mapbox的token


	var t = time.Now() // get current time
	errDir := fmt.Sprintf("%s/error.output", rootDir)
	logfile,_ := os.Create(errDir)
	//logfile,err:=os.OpenFile(errDir,os.O_RDWR|os.O_CREATE,0666)
	defer logfile.Close()
	if err!=nil{
		fmt.Printf("%s\r\n",err.Error())
		os.Exit(-1)
	}
	for level := currentLevel; level < maxLevel; level++ {
		levelDir := fmt.Sprintf("%s/%d", rootDir, level)
		err := os.Mkdir(levelDir, os.ModePerm)
		if err != nil {
			fmt.Fprintln(logfile,err)
			log.Fatal(err)
		}
		//for x := int(math.Pow(2, float64(level-minLevel))) * minX; x < int(math.Pow(2, float64(level-minLevel))) * (minX + 2); x++ {
		for x := int(math.Pow(2, float64(level-minLevel))) * minX; x < int(math.Pow(2, float64(level-minLevel))) * maxX; x++ {
			xDir := fmt.Sprintf("%s/%d/%d", rootDir, level, x)
			err := os.Mkdir(xDir, os.ModePerm)
			if err != nil {
				fmt.Fprintln(logfile,err)
				log.Fatal(err)
			}
			//for y := int(math.Pow(2, float64(level-minLevel))) * minY; y <= int(math.Pow(2, float64(level-minLevel))) * (minY+2); y++ {
			for y := int(math.Pow(2, float64(level-minLevel))) * minY; y < int(math.Pow(2, float64(level-minLevel))) * maxY; y++ {
				name := fmt.Sprintf("%d/%d/%d", level, x, y)
				url := fmt.Sprintf("正在下载%s%s.vector.pbf?access_token=%s", preUrl, name, mapboxToken)
				fmt.Println(url)
				result := fmt.Sprintf("%s/%d/%d/%d.pbf", rootDir, level, x, y)
				wg.Add(1)
				go requestInfo(url, result, &wg)
			}
		}
	}
	elapsed := time.Since(t)
	fmt.Println("下载完成！耗时: ", elapsed)
}
