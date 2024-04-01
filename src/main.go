package main

import (
	"encoding/json"
	"fmt"
	"go-colly-lib/src/crawler"
	"os"

	"github.com/gocolly/colly"
)

func main() {
	saveData()
}

func saveData() {
	result := crawler.GetEpicGameData()

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
