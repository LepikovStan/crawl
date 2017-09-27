package main

import (
	"fmt"
	"time"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"io"
	"strings"
	"strconv"
	"flag"
	//"sync"
	"os"
	"bytes"
	"sync"
)

var refsCount int

func getInitUrls() []string {
	sitesList := []string{
		//"http://calm.com",
		//"http://donothingfor2minutes.com/",
		//"https://hosting.reg.ru/",
		//"http://stenadobra.ru/",
		"http://humandescent.com",
		 //"http://thefirstworldwidewebsitewerenothinghappens.com",
		// "http://button.dekel.ru",
		// "http://www.randominio.com/",
		// "http://thenicestplaceontheinter.net/",
		// "http://www.catsthatlooklikehitler.com/",
		// "http://www.thefirstworldwidewebsitewerenothinghappens.com/",
		// "http://www.donothingfor2minutes.com/",
		// "http://www.howmanypeopleareinspacerightnow.com/",
		// "http://www.humanclock.com/",
		// "http://fucking-great-advice.ru/",
		// "http://www.cesmes.fi/pallo.swf",
		// "http://button.dekel.ru/",
		// "http://www.rainfor.me/",
		// "http://loudportraits.com/",
		// "http://sprosimamu.ru/",
		// "http://www.bandofbridges.com/",
		// "http://www.catsboobs.com/",
		// "http://www.incredibox.com/",
	}
	return sitesList
}

type Ref struct {
	Title string
	Href string
}

type Backlink struct {
	Url string
	Body io.Reader
}

func crawler(urls chan string, body chan Backlink) {
	for {
		url := <- urls
		fmt.Println("start crawl", url)

		resp, err := http.Get(url)
		if (err != nil) {
			fmt.Println("crawler error", err)
			//errors <- err
		}

		if resp.StatusCode == 200 {
			backlink := Backlink{
				Url: url,
				Body: resp.Body,
			}
		    body <- backlink
		}
	}
}

func setToQueue (urls chan string, urlList[]string) {
	for _, url := range urlList {
		urls <- url
	}
}

var workersCount int
var maxDepth int
var currentDepth int
func initFlags() {
	flag.IntVar(&workersCount, "maxWorkers", 1, "")
	flag.IntVar(&maxDepth, "maxDepth", 1, "")
	flag.Parse()
}


func parser(
	urls chan string,
	body chan Backlink,
	result chan string,
	currentDepth *int,
	maxDepth int,
	wg *sync.WaitGroup,
) {
	for {
		fmt.Println("parser")
		backlink := <- body

		doc, err := goquery.NewDocumentFromReader(backlink.Body)
		urlsList := []string{}

		if err != nil {
			fmt.Println("parser error", err)
		}


		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			Href, _ := s.Attr("href")
			if (!strings.HasPrefix(Href, "http")) {
				return
			}
			// Title := strings.TrimSpace(s.Text())
			urlsList = append(urlsList, Href)

			fmt.Println("   ", backlink.Url, "->", Href)
			var buffer bytes.Buffer
			buffer.WriteString(backlink.Url)
			buffer.WriteString(" -> ")
			buffer.WriteString(Href)
			buffer.WriteString("\n")
			result <- buffer.String()
			refsCount++
		})
		if (*currentDepth < maxDepth) {
			*currentDepth++
			go setToQueue(urls, urlsList)
		} else {
			fmt.Println("Depth is end", refsCount)
			//wg.Done()
		}
	}
}

func writer(result chan string, resultFile *os.File, mx *sync.RWMutex) {
	for {
		mx.Lock()
		if _, err := resultFile.WriteString(<- result); err != nil {
			fmt.Println(err)
		}
		mx.Unlock()
	}
}

func main() {
	fmt.Println("Start...")
	start := time.Now()


	var wg sync.WaitGroup
	var buffer bytes.Buffer
	urls := make(chan string)
	body := make(chan Backlink)
	result := make(chan string)

	buffer.WriteString("result.")
	buffer.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
	buffer.WriteString(".log")
	resultFile, err := os.OpenFile(buffer.String(), os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(err)
	}
	//backlinks := make(chan BacklinksList)
	//errors := make(chan error)
	initFlags()
	//wgCounter := 2*workersCount+1
	//fmt.Println("wgCounter", wgCounter)
	//
	urlsList := getInitUrls()
	currentDepth = 1

	wg.Add(maxDepth)
	go setToQueue(urls, urlsList)
	for i:=0;i<workersCount;i++ {
		go crawler(urls, body)
		go parser(urls, body, result, &currentDepth, maxDepth, &wg)
	}
	mx := sync.RWMutex{}
	go writer(result, resultFile, &mx)
	//wg.Wait()
	//fmt.Println("end")
	//
	//
	end := time.Now()
	fmt.Println("\n")
	fmt.Println(end.Sub(start), "refsCount: ", refsCount)

	var input string
	fmt.Scanln(&input)
}
