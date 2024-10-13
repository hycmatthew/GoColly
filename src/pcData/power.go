package pcData

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type PowerSpec struct {
	Code        string
	Brand       string
	ReleaseDate string
	Wattage     string
	Size        string
	Modular     string
	Efficiency  string
	Length      string
	PriceUS     string
	PriceHK     string
	PriceCN     string
	LinkUS      string
	LinkHK      string
	LinkCN      string
	Img         string
}

type PowerType struct {
	Brand       string
	ReleaseDate string
	Wattage     string
	Size        string
	Modular     string
	Efficiency  string
	Length      string
	PriceUS     string
	PriceHK     string
	PriceCN     string
	LinkUS      string
	LinkHK      string
	LinkCN      string
	Img         string
}

func GetPowerSpec(record LinkRecord) PowerSpec {

	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"www.newegg.com",
			"newegg.com",
			"www.price.com.hk",
			"price.com.hk",
			"product.pconline.com.cn",
			"pconline.com.cn",
			"pangoly.com",
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	specCollector := collector.Clone()

	ssdData := getPowerSpecData(record.LinkSpec, specCollector)
	ssdData.Code = record.Name
	ssdData.Brand = record.Brand
	ssdData.LinkCN = record.LinkCN
	if ssdData.LinkUS == "" {
		ssdData.LinkUS = record.LinkUS
	}
	ssdData.LinkHK = record.LinkHK
	if record.PriceCN != "" {
		ssdData.PriceCN = record.PriceCN
	}
	return ssdData
}

func getPowerSpecData(link string, collector *colly.Collector) PowerSpec {
	releaseDate := ""
	wattage := ""
	size := ""
	modular := ""
	efficiency := ""
	length := ""
	price := ""
	usLink := ""
	imgLink := ""

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".tns-inner img", "src")
		loopBreak := false

		element.ForEach("table.table-prices tr", func(i int, item *colly.HTMLElement) {
			if !loopBreak {
				price = extractFloatStringFromString(item.ChildText(".detail-purchase strong"))
				tempLink := item.ChildAttr(".detail-purchase", "href")

				if strings.Contains(tempLink, "amazon") {
					amazonLink := strings.Split(tempLink, "?tag=")[0]
					usLink = amazonLink
					loopBreak = true
				}
				if strings.Contains(tempLink, "newegg") {
					neweggLink := strings.Split(tempLink, "url=")[1]
					UnescapeLink, _ := url.QueryUnescape(neweggLink)
					usLink = strings.Split(UnescapeLink, "\u0026")[0]
					loopBreak = true
				}
			}
		})

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Release Date":
				releaseDate = item.ChildText("td span")
			case "Capacity":
				wattage = item.ChildTexts("td")[1]
			case "Interface":
				size = item.ChildTexts("td")[1]
			case "Form Factor":
				modular = item.ChildTexts("td")[1]
			case "NAND Flash Type":
				efficiency = item.ChildTexts("td")[1]
			case "Max Sequential Read":
				length = item.ChildTexts("td")[1]
			}
		})
	})

	collector.Visit(link)

	return PowerSpec{
		ReleaseDate: releaseDate,
		Wattage:     wattage,
		Size:        size,
		Modular:     modular,
		Efficiency:  efficiency,
		Length:      length,
		PriceUS:     price,
		LinkUS:      usLink,
		Img:         imgLink,
	}
}

func getPowerUSPrice(link string, collector *colly.Collector) float64 {
	price := 0.0

	collectorErrorHandle(collector, link)
	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current")), 64); err == nil {
			price = s
		}
	})

	collector.Visit(link)
	return price
}

func getPowerHKPrice(link string, collector *colly.Collector) float64 {
	price := 0.0

	collectorErrorHandle(collector, link)

	collector.OnHTML(".line-05", func(element *colly.HTMLElement) {

		element.ForEach(".product-price", func(i int, item *colly.HTMLElement) {
			fmt.Println(extractFloatStringFromString(element.ChildText("span")))
			if price == 0.0 {
				if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText("span")), 64); err == nil {
					price = s
					//fmt.Println(price)
				} else {
					fmt.Println(err)
				}
			}
		})
	})

	collector.Visit(link)
	return price
}

func getPowerCNPrice(link string, collector *colly.Collector) float64 {
	price := 0.0

	collectorErrorHandle(collector, link)

	collector.OnHTML(".product-mallSales", func(element *colly.HTMLElement) {
		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText("em.price")), 64); err == nil {
			price = s
			// fmt.Println(price)
		} else {
			fmt.Println(err)
		}
	})

	collector.Visit(link)
	return price
}
