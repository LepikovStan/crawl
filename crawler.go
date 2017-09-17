package main

import (
	"fmt"
	"time"
	"net/http"
	"sync"
	// "io/ioutil"
	// "html"
	"github.com/PuerkitoBio/goquery"
	// "github.com/moovweb/gokogiri"
	"io"
	"strings"
)

func getSitesList() []string {
	sitesList := []string{
		"http://donothingfor2minutes.com/",
		// "http://stenadobra.ru/",
		// "http://humandescent.com",
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
	defer wg.Done()
	url := <- urls
	fmt.Println("start crawl", url)

	resp, err := http.Get(url)
	if (err != nil) {
		errors <- err
	}

	if resp.StatusCode == 200 {
	    body <- resp.Body
	}
}

func parser(body chan io.Reader, backlinks chan BacklinksList, errors chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("start parse")

	doc, err := goquery.NewDocumentFromReader(<- body)
 	if err != nil {
		fmt.Println(err)
 		errors <- err
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		Title := strings.TrimSpace(s.Text())
		Href, _ := s.Attr("href")

		fmt.Println("   ", Title, ":", Href)
	})
}

func getFromQueue(urls chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	sitesList := getSitesList()
	fmt.Println("getFromQueue")

	for _, url := range sitesList {
		urls <- url
	}
}

func main() {
	fmt.Println("Start...")
	start := time.Now()


	var wg sync.WaitGroup
	urls := make(chan string)
	body := make(chan io.Reader)
	backlinks := make(chan BacklinksList)
	errors := make(chan error)

	wg.Add(3)
	go getFromQueue(urls, &wg)
	go crawler(urls, body, errors, &wg)
	go parser(body, backlinks, errors, &wg)
	wg.Wait()
	fmt.Println("end")


	end := time.Now()
	fmt.Println("\n")
	fmt.Println(end.Sub(start))
}
