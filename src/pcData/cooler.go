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

type CoolerSpec struct {
	Code           string
	Brand          string
	ReleaseDate    string
	Sockets        []string
	IsLiquidCooler string
	Size           string
	NoiseLevel     string
	FanSpeed       string
	PriceUS        string
	PriceHK        string
	PriceCN        string
	LinkUS         string
	LinkHK         string
	LinkCN         string
	Img            string
}

type CoolerType struct {
	Brand          string
	ReleaseDate    string
	Sockets        []string
	IsLiquidCooler string
	Size           string
	NoiseLevel     string
	FanSpeed       string
	PriceUS        string
	PriceHK        string
	PriceCN        string
	LinkUS         string
	LinkHK         string
	LinkCN         string
	Img            string
}

func GetCoolerSpec(record LinkRecord) CoolerSpec {

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

	powerData := getCoolerSpecData(record.LinkSpec, specCollector)
	powerData.Code = record.Name
	powerData.Brand = record.Brand
	powerData.LinkCN = record.LinkCN
	if powerData.LinkUS == "" {
		powerData.LinkUS = record.LinkUS
	}
	powerData.LinkHK = record.LinkHK
	if record.PriceCN != "" {
		powerData.PriceCN = record.PriceCN
	}
	return powerData
}

func getCoolerSpecData(link string, collector *colly.Collector) CoolerSpec {
	releaseDate := ""
	var socketslist []string
	isLiquidCooler := ""
	size := ""
	noiseLevel := ""
	fanSpeed := ""
	price := ""
	usLink := ""
	imgLink := ""

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".tns-inner .tns-item img", "src")
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
			case "Supported Sockets":
				item.ForEach("td li", func(i int, subitem *colly.HTMLElement) {
					socketslist = append(socketslist, subitem.Text)
				})
			case "Liquid Cooler":
				isLiquidCooler = item.ChildTexts("td")[1]
			case "Radiator Size":
				size = item.ChildTexts("td")[1]
			case "Noise Level":
				noiseLevel = item.ChildTexts("td")[1]
			case "Fan RPM":
				fanSpeed = item.ChildTexts("td")[1]
			}
		})
	})

	collector.Visit(link)

	return CoolerSpec{
		ReleaseDate:    releaseDate,
		Sockets:        socketslist,
		IsLiquidCooler: isLiquidCooler,
		Size:           size,
		NoiseLevel:     noiseLevel,
		FanSpeed:       fanSpeed,
		PriceUS:        price,
		LinkUS:         usLink,
		Img:            imgLink,
	}
}

func getCoolerUSPrice(link string, collector *colly.Collector) float64 {
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

func getCoolerHKPrice(link string, collector *colly.Collector) float64 {
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

func getCoolerCNPrice(link string, collector *colly.Collector) float64 {
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
