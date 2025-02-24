package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type CSVData struct {
	brand    string
	name     string
	specLink string
}

func createTempCSV(data []CSVData) {
	file, err := os.Create("res/output.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// 創建 CSV 寫入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 寫入標題行
	err = writer.Write([]string{"brand", "name", "specLink"})
	if err != nil {
		fmt.Println("Error writing header:", err)
		return
	}

	// 寫入每一行數據
	for _, item := range data {
		err := writer.Write([]string{
			item.brand,
			item.name,
			item.specLink,
		})
		if err != nil {
			fmt.Println("Error writing record:", err)
			return
		}
	}

	fmt.Println("Data successfully written to output.csv")
}

func main() {
	// 打開文件
	file, err := os.Open("res/webLink.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// 創建一個切片來存儲每一行
	var lines []string

	// 使用 bufio 逐行讀取
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// 檢查是否出現錯誤
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	getDatgaLogic(lines)
}

func getDatgaLogic(data []string) {
	timeSet := 5000
	extraTry := 50
	maxRetryTime := 3
	retryTime := 0
	timeDuration := time.Duration(timeSet) * time.Millisecond
	ticker := time.NewTicker(timeDuration)

	var List []CSVData
	count := 0

	go func() {
		for {
			<-ticker.C
			spec := data[count]
			record := GetLinkData(spec)
			if len(record) > 0 || retryTime == maxRetryTime {
				List = append(List, record...)
				retryTime = 0
				count++
			} else {
				retryTime++
			}

			if count == len(data) {
				createTempCSV(List)
				ticker.Stop()
				runtime.Goexit()
			}
		}
	}()

	listLen := time.Duration(timeSet * (len(data) + extraTry))
	time.Sleep(time.Second * listLen)

}

func GetLinkData(link string) []CSVData {

	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"pangoly.com",
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	linkData := getWebpageLinkData(link, collector)
	return linkData
}

func getWebpageLinkData(link string, collector *colly.Collector) []CSVData {
	fmt.Println("Link: ", link)
	var csvlist []CSVData
	collectorErrorHandle(collector, link)

	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		brand := element.ChildText(".breadcrumb .active")
		// tempData := element.ChildText(".products-grid")

		element.ForEach(".products-grid .productItem", func(i int, item *colly.HTMLElement) {
			removedText := item.DOM.Find(".productItemLink header").Children().Remove().End().Text()
			tempName := strings.TrimSpace(removedText)
			tempLink := item.ChildAttr(".productItemLink", "href")
			price := item.ChildText(".price .amprice")
			updatedName := strings.TrimSpace(RemoveBrandsFromName(brand, tempName))

			// name is not completed
			if strings.Contains(updatedName, "...") {
				fmt.Println("name: ", updatedName)
				nameFromUrl := GetLastSegment(tempLink)
				typeFromUrl := ExtractTypeFromURL(tempLink)
				newName := replaceHyphensAndCapitalize(typeFromUrl, nameFromUrl)
				updatedName = strings.TrimSpace(RemoveBrandsFromName(brand, newName))
			}

			if price != "" {
				fmt.Println("price: ", price)
				csvItem := CSVData{
					brand:    brand,
					name:     transformString(updatedName),
					specLink: tempLink,
				}
				csvlist = append(csvlist, csvItem)
			}
		})
	})
	/*
		collector.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
			fmt.Printf(e.Text)
		})
	*/
	collector.Visit(link)

	return csvlist
}

func collectorErrorHandle(collector *colly.Collector, link string) {
	collector.OnRequest(func(r *colly.Request) {
		// USER_AGENT = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36'
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
	})

	collector.OnError(func(response *colly.Response, err error) {
		fmt.Println("请求期间发生错误,则调用:", err, " - link: ", link)
	})

	collector.OnResponse(func(response *colly.Response) {
		fmt.Println("收到响应后调用:", response.Request.URL)
	})
}

func transformString(input string) string {
	// 正則表達式來匹配 2x16gb, 4 x 16gb, 2x 32gb 等格式
	re := regexp.MustCompile(`(\d+)\s*[xX]\s*(\d+[gG][bB])`)

	// 使用 ReplaceAllStringFunc 來處理匹配到的部分
	transformed := re.ReplaceAllStringFunc(input, func(match string) string {
		// 將 match 中的空白和大小寫統一處理
		match = strings.ReplaceAll(match, "X", "x")
		match = strings.ReplaceAll(match, " ", "")
		return "(" + match + ")"
	})

	transformed = strings.ReplaceAll(transformed, "((", "(")
	transformed = strings.ReplaceAll(transformed, "))", ")")
	return transformed
}

func replaceHyphensAndCapitalize(partType string, s string) string {
	fmt.Println("type: ", partType)
	keepCaps := []string{"rgb", "amd", "gb", "cl", "ddr", "lpx"}
	// Replace hyphens with spaces
	s = strings.ReplaceAll(s, "-", " ")

	// Split the string into words
	words := strings.Fields(s)

	// Capitalize the first letter of each word
	for i, word := range words {
		isConverted := false
		/*
			if strings.Contains(word, "gb") && strings.Contains(word, "x") {
				words[i] = "(" + words[i] + ")"
			}
		*/
		for _, keep := range keepCaps {
			if strings.Contains(word, keep) {
				words[i] = strings.ReplaceAll(word, keep, strings.ToUpper(keep))
				isConverted = true
				break
			}
		}

		if strings.Contains(word, "mhz") {
			words[i] = strings.ReplaceAll(word, "mhz", "MHz")
		}

		if len(word) > 0 && !isConverted {
			words[i] = string(unicode.ToUpper(rune(word[0]))) + word[1:]
		}
	}
	// Join the words back into a single string
	return strings.Join(words, " ")
}
