package pcData

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type SsdRecord struct {
	Name   string
	LinkCN string
	LinkHK string
	LinkUS string
}

type SsdType struct {
	Brand    string
	Series   string
	Model    string
	Capacity string
	Speed    string
	Timing   string
	Voltage  string
	Channel  string
	Profile  string
	PriceUS  float64
	PriceHK  float64
	PriceCN  float64
	Img      string
}

func GetSsdData(cnLink string, enLink string, hkLink string) SsdType {

	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			// "https://nanoreview.net",
			"nanoreview.net",
			"www.newegg.com",
			"newegg.com",
			// "https://cu.manmanbuy.com",
			"cu.manmanbuy.com",
			"www.price.com.hk",
			"price.com.hk",
			"detail.zol.com.cn",
			"zol.com.cn",
			"product.pconline.com.cn",
			"pconline.com.cn",
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	usCollector := collector.Clone()
	cnCollector := collector.Clone()
	// hkCollector := collector.Clone()

	ramData := getSsdUSPrice(enLink, usCollector)
	ramData.PriceCN = getSsdCNPrice(cnLink, cnCollector)
	// ramData.PriceHK = getSsdHKPrice(hkLink, hkCollector)

	return ramData
}

func getSsdUSPrice(link string, collector *colly.Collector) SsdType {
	brand := ""
	series := ""
	model := ""
	capacity := ""
	speed := ""
	timing := ""
	voltage := ""
	channel := ""
	profile := ""
	price := 0.0
	imgLink := ""

	collectorErrorHandle(collector, link)
	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")

		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current")), 64); err == nil {
			price = s
			fmt.Println(price)
		}

		element.ForEach(".tab-box .tab-panes tr", func(i int, item *colly.HTMLElement) {
			fmt.Println(item.ChildText("td"))
			switch item.ChildText("th") {
			case "Brand":
				brand = item.ChildText("td")
			case "Series":
				series = item.ChildText("td")
			case "Model":
				model = item.ChildText("td")
			case "Capacity":
				capacity = item.ChildText("td")
			case "Speed":
				speed = item.ChildText("td")
			case "Timing":
				timing = item.ChildText("td")
			case "Voltage":
				voltage = item.ChildText("td")
			case "Multi-channel Kit":
				channel = item.ChildText("td")
			case "BIOS/Performance Profile":
				profile = item.ChildText("td")
			}
		})
	})

	collector.Visit(link)

	return SsdType{
		Brand:    brand,
		Series:   series,
		Model:    model,
		Capacity: capacity,
		Speed:    speed,
		Timing:   timing,
		Voltage:  voltage,
		Channel:  channel,
		Profile:  profile,
		PriceUS:  price,
		PriceHK:  0,
		PriceCN:  0,
		Img:      imgLink,
	}
}

func getSsdHKPrice(link string, collector *colly.Collector) float64 {
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

func getSsdCNPrice(link string, collector *colly.Collector) float64 {
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
