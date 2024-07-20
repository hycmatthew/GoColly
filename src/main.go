package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"go-colly-lib/src/crawler"
	"go-colly-lib/src/pcData"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
	// saveData()
	udpateCPULogic()
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

func saveData() {
	result := crawler.GetEpicGameData()
	fmt.Println(result)

	jsonData, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write JSON data to file
	err = os.WriteFile("tmp/epicGame.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func udpateCPULogic() {
	dataList := readCsvFile("res/cpudata.csv")
	var recordList []pcData.CPURecord

	for i := 1; i < len(dataList); i++ {
		data := dataList[i]
		record := pcData.CPURecord{Name: data[0], LinkSpec: data[1], LinkCN: data[2], LinkUS: data[3], LinkHK: data[4]}
		recordList = append(recordList, record)
	}

	ticker := time.NewTicker(1500 * time.Millisecond)
	count := 0

	go func() {
		for {
			<-ticker.C

			pcData.GetCPUData(recordList[count].LinkSpec, recordList[count].LinkUS, recordList[count].LinkCN, recordList[count].LinkHK)
			count++
			if count == 2 {
				ticker.Stop()
				runtime.Goexit()
			}
		}
	}()

	listLen := time.Duration(len(recordList) * 2)
	time.Sleep(time.Second * listLen)
}

func getCollyData() {
	// Create a new Colly collector
	c := colly.NewCollector(
		colly.AllowedDomains("store.epicgames.com", "epicgames.com", "www.cpu-monkey.com"),
	)

	// On every a element which has href attribute call callback
	c.OnHTML("a", func(e *colly.HTMLElement) {
		link := e.Text
		// Print link
		fmt.Printf("123123")
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		// c.Visit(e.Request.AbsoluteURL(link))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping
	c.Visit("https://store-site-backend-static-ipv4.ak.epicgames.com/freeGamesPromotions?locale=zh-Hant&country=HK&allowCountries=HK")
}
