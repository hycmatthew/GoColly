package main

import (
	"fmt"
	"go-colly-lib/src/crawler"

	"github.com/gocolly/colly"
)

func main() {

	// getCollyData()
	result := crawler.GetEpicGameData()

	for _, item := range result {
		fmt.Printf("Name: %s, StartDate: %s\n", item.Name, item.StartDate)
	}
}

func saveData() {

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
