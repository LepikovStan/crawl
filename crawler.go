//package main
//
//import (
//	"fmt"
//	"time"
//)
//
//func pinger(c chan int) {
//	for i := 0; i<5; i++ {
//		c <-  i
//	}
//}
//func printer(c chan int) {
//	for {
//		msg := <- c
//		fmt.Println(msg)
//		if msg == 3 {
//			go func () {
//				c <- msg
//			}()
//		}
//		time.Sleep(time.Second * 1)
//		//c <- m
//		//time.Sleep(time.Second * 1)
//	}
//}
//
//func printer2(d chan int) {
//	for {
//		msg := <- d
//		fmt.Println(msg)
//		//time.Sleep(time.Second * 1)
//	}
//}
//func ponger(c chan int) {
//	for i := 0; i<5; i++ {
//		c <- i
//	}
//}
//func main() {
//	var c chan int = make(chan int)
//
//	go pinger(c)
//	for i:=0;i<1;i++ {
//		go printer(c)
//	}
//
//	var input string
//	fmt.Scanln(&input)
//}

package main

import (
	"fmt"
	"time"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"io"
	"strings"
	"flag"
	"sync"
	"os"
	"bytes"
)

var refsCount int

func getInitUrls() []string {
	sitesList := []string{
		//"http://calm.com",
		//"http://donothingfor2minutes.com/",
		//"https://hosting.reg.ru/",
		//"http://stenadobra.ru/",
		//"http://humandescent.com",
		 "http://thefirstworldwidewebsitewerenothinghappens.com",
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

type BacklinksList struct {
	Url string
	Links []Ref
}

func crawler(urls chan string, body chan io.Reader) {
	for {
		url := <- urls
		fmt.Println("start crawl", url)

		resp, err := http.Get(url)
		if (err != nil) {
			fmt.Println(err)
			//errors <- err
		}

		if resp.StatusCode == 200 {
		    body <- resp.Body
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
	body chan io.Reader,
	result chan string,
	currentDepth *int,
	maxDepth int,
	wg *sync.WaitGroup,
) {
	for {
		fmt.Println("parser")
		siteBody := <- body
		doc, err := goquery.NewDocumentFromReader(siteBody)
		urlsList := []string{}

		if err != nil {
			fmt.Println(err)
		}


		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			Href, _ := s.Attr("href")
			if (!strings.HasPrefix(Href, "http")) {
				return
			}
			// Title := strings.TrimSpace(s.Text())
			urlsList = append(urlsList, Href)

			// fmt.Println("   ", Title, ":", Href)
			// result <-
			refsCount++
		})
		if (*currentDepth < maxDepth) {
			*currentDepth++
			go setToQueue(urls, urlsList)
		} else {
			fmt.Println("Depth is end", refsCount)
			wg.Done()
		}
	}
}

func writer(result chan string, resultFile *os.File) {
	for {
		var buffer bytes.Buffer
		resultLine := <- result
		buffer.WriteString(resultLine)
		buffer.WriteString("\n")
		if _, err := resultFile.WriteString(buffer.String()); err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	fmt.Println("Start...")
	start := time.Now()


	var wg sync.WaitGroup
	urls := make(chan string)
	body := make(chan io.Reader)
	result := make(chan string)
	resultFile, err := os.OpenFile("result.log", os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(err)
	}
	backlinksList = new(BacklinksList)
	//backlinks := make(chan BacklinksList)
	//errors := make(chan error)
	initFlags()
	//wgCounter := 2*workersCount+1
	//fmt.Println("wgCounter", wgCounter)
	//
	wg.Add(5)
	urlsList := getInitUrls()
	currentDepth = 1

	go setToQueue(urls, urlsList)
	for i:=0;i<workersCount;i++ {
		go crawler(urls, body)
		go parser(urls, body, result, &currentDepth, maxDepth, &wg)
	}
	go writer(result, resultFile)
	wg.Wait()
	//fmt.Println("end")
	//
	//
	end := time.Now()
	fmt.Println("\n")
	fmt.Println(end.Sub(start), "refsCount: ", refsCount)

	// var input string
	// fmt.Scanln(&input)
}
