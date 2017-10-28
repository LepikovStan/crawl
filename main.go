package main

import (
	"fmt"
	"time"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"io"
	"strings"
	"flag"
	"os"
	"sync"
	"bufio"
)

var refsCount int

type Ref struct {
	Title string
	Href string
}

type Backlink struct {
	Url string
	Body io.Reader
}

func crawl(urls chan string, body chan Backlink) {
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

func setToQueue (urls chan string, urlList[]string, refsToCrawlCount *int) {
	*refsToCrawlCount = *refsToCrawlCount + len(urlList)
	fmt.Println("setToQueue", *refsToCrawlCount, len(urlList))
	for _, url := range urlList {
		urls <- url
	}
}

var parsersCount int
var crawlersCount int
var maxDepth int
var currentDepth int
var refsToCrawlCount int
func initFlags() {
	flag.IntVar(&parsersCount, "parsers", 1, "")
	flag.IntVar(&crawlersCount, "crawlers", 1, "")
	flag.IntVar(&maxDepth, "depth", 1, "")
	flag.Parse()
}

func parse(
	urls chan string,
	body chan Backlink,
	result chan string,
	currentDepth *int,
	maxDepth int,
	wg *sync.WaitGroup,
	refsToCrawlCount *int,
) {
	for {
		backlink := <- body

		doc, err := goquery.NewDocumentFromReader(backlink.Body)
		var urlsList []string

		if err != nil {
			fmt.Println("parser error", err)
		}


		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			Href, _ := s.Attr("href")
			if (!strings.HasPrefix(Href, "http")) {
				return
			}
			urlsList = append(urlsList, Href)

			//fmt.Println("   ", backlink.Url, "->", Href)
			result <- fmt.Sprintf("%s -> %s\n", backlink.Url, Href)
			refsCount++
		})
		*refsToCrawlCount--

		if (*currentDepth < maxDepth) {
			*currentDepth++
			go setToQueue(urls, urlsList, refsToCrawlCount)
		} else {
			if (*refsToCrawlCount == 1) {
				fmt.Println("All done")
				wg.Done()
			}
		}
	}
}

type Writer struct {
	mux sync.Mutex
}

func (w *Writer) write(resultFile *os.File, s string) {
	w.mux.Lock()
	defer w.mux.Unlock()
	if _, err := resultFile.WriteString(s); err != nil {
		fmt.Println(err)
	}
}

func writer(result chan string, resultFile *os.File, fileWriter Writer) {
	for {
		fileWriter.write(resultFile, <- result)
	}
}

func readFile(path string) []string {
	var result []string
	inFile, _ := os.Open(path)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	return result
}

// http://humandescent.com
func main() {
	fmt.Println("Start...")
	start := time.Now()


	var wg sync.WaitGroup
	urls := make(chan string)
	body := make(chan Backlink)
	result := make(chan string)

	fileName := fmt.Sprintf("result.%d.log", time.Now().Unix())
	resultFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(err)
	}

	initFlags()

	urlsList := readFile("input.txt")
	currentDepth = 1
	refsToCrawlCount = 0

	wg.Add(1)
	go setToQueue(urls, urlsList, &refsToCrawlCount)
	for i:=0;i<crawlersCount;i++ {
		go crawl(urls, body)
	}
	for i:=0;i<parsersCount;i++ {
		go parse(urls, body, result, &currentDepth, maxDepth, &wg, &refsToCrawlCount)
	}
	fileWriter := Writer{}
	go writer(result, resultFile, fileWriter)
	wg.Wait()

	end := time.Now()
	fmt.Println("\n")
	fmt.Println(end.Sub(start), "refsCount: ", refsCount)
}
