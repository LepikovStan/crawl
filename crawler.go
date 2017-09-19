package main

import (
	"fmt"
	"time"
	"net/http"
	"sync"
	"github.com/PuerkitoBio/goquery"
	"io"
	"strings"
	"flag"
	//"sync"
)

var refsCount int

func getSitesList() []string {
	sitesList := []string{
		"http://donothingfor2minutes.com/",
		"http://stenadobra.ru/",
		"http://humandescent.com",
		// "http://thefirstworldwidewebsitewerenothinghappens.com",
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

func crawler(urls chan string, body chan io.Reader, errors chan error, wg *sync.WaitGroup) {
	for {
		defer fmt.Println("ae1")
		url := <- urls
		fmt.Println("start crawl", url)

		resp, err := http.Get(url)
		if (err != nil) {
			errors <- err
		}

		if resp.StatusCode == 200 {
		    body <- resp.Body
		}
		wg.Done()
	}
}

func parser(body chan io.Reader, backlinks chan BacklinksList, errors chan error, wg *sync.WaitGroup) {
	for {
		doc, err := goquery.NewDocumentFromReader(<- body)
	 	if err != nil {
			fmt.Println(err)
	 		errors <- err
		}

		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			Href, _ := s.Attr("href")
			if (!strings.HasPrefix(Href, "http")) {
				return
			}
			Title := strings.TrimSpace(s.Text())

			fmt.Println("   ", Title, ":", Href)
			refsCount++
		})
		wg.Done()
	}
}

func getFromQueue(index int, urls chan string, wg *sync.WaitGroup) {
	// defer wg.Done()
	sitesList := getSitesList()
	// fmt.Println("getFromQueue")

	for _, url := range sitesList {
		urls <- url
		RemoveFromQueue(sitesList, index)
		wg.Done()
	}

	// urls <- sitesList[index]
	// RemoveFromQueue(sitesList, index)
}

func RemoveFromQueue(s []string, index int) []string {
    return append(s[:index], s[index+1:]...)
}

var workersCount int
func initFlags() {
	flag.IntVar(&workersCount, "workersCount", 1, "")
	flag.Parse()
}

func main() {
	fmt.Println("Start...")
	start := time.Now()


	var wg sync.WaitGroup
	urls := make(chan string)
	body := make(chan io.Reader)
	backlinks := make(chan BacklinksList)
	errors := make(chan error)
	sitesList := getSitesList()
	sitesListLength := len(sitesList)
	initFlags()

	wg.Add(sitesListLength*3*workersCount)
	for i:=0;i<workersCount;i++ {
		go getFromQueue(0, urls, &wg)
		go crawler(urls, body, errors, &wg)
		go parser(body, backlinks, errors, &wg)
	}
	wg.Wait()
	fmt.Println("end")


	end := time.Now()
	fmt.Println("\n")
	fmt.Println(end.Sub(start), "refsCount: ", refsCount)
}
