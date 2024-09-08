package pcData

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type MotherboardRecord struct {
	Name   string
	LinkCN string
	LinkHK string
	LinkUS string
}

type MotherboardType struct {
	Name        string
	Brand       string
	Socket      string
	Chipset     string
	Ram         string
	RamType     string
	MemorySlots string
	Storage     string
	FormFactor  string
	PriceUS     float64
	PriceHK     float64
	PriceCN     float64
	Img         string
}

func GetMotherboardData(cnLink string, enLink string, hkLink string) MotherboardType {

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

	motherboardData := getMotherboardUSPrice(enLink, usCollector)
	motherboardData.PriceCN = getMotherboardCNPrice(cnLink, cnCollector)
	// motherboardData.PriceHK = getMotherboardHKPrice(hkLink, hkCollector)

	return motherboardData
}

func getMotherboardUSPrice(link string, collector *colly.Collector) MotherboardType {
	name := ""
	brand := ""
	socket := ""
	chipset := ""
	ram := ""
	ramType := ""
	storage := ""
	formFactor := ""
	memorySlots := ""
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
			case "Model":
				name = item.ChildText("td")
			case "CPU Socket Type":
				socket = item.ChildText("td")
			case "Chipset":
				chipset = item.ChildText("td")
			case "Number of Memory Slots":
				memorySlots = item.ChildText("td")
				ramType = item.ChildText("td")
			case "Memory Standard":
				ram = item.ChildText("td")
			case "M.2":
				storage = item.ChildText("td")
			case "Form Factor":
				formFactor = item.ChildText("td")
			}
		})
	})

	collector.Visit(link)

	fmt.Println(MotherboardType{
		Name:        name,
		Brand:       brand,
		Socket:      socket,
		Chipset:     chipset,
		Ram:         ram,
		RamType:     ramType,
		MemorySlots: memorySlots,
		Storage:     storage,
		FormFactor:  formFactor,
		PriceUS:     price,
		PriceHK:     0,
		PriceCN:     0,
		Img:         imgLink,
	})

	return MotherboardType{
		Name:        name,
		Brand:       brand,
		Socket:      socket,
		Chipset:     chipset,
		Ram:         ram,
		RamType:     ramType,
		MemorySlots: memorySlots,
		Storage:     storage,
		FormFactor:  formFactor,
		PriceUS:     price,
		PriceHK:     0,
		PriceCN:     0,
		Img:         imgLink,
	}
}

func getMotherboardHKPrice(link string, collector *colly.Collector) float64 {
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

func getMotherboardCNPrice(link string, collector *colly.Collector) float64 {
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
