package pcData

import (
	"fmt"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type GPURecord struct {
	Name   string
	LinkCN string
	LinkHK string
	LinkUS string
}

type GPUType struct {
	Name       string
	Brand      string
	MemorySize string
	MemoryType string
	MemoryBus  string
	Clock      string
	Power      int
	Length     int
	Slot       string
	Width      int
	PriceUS    float64
	PriceHK    float64
	PriceCN    float64
	Img        string
}

func GetGPUSpec(name string, link string) GPUSpecTempStruct {
	fakeChrome := req.C().ImpersonateChrome().SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36").SetTLSFingerprintChrome()

	fmt.Println(fakeChrome.Headers.Get("user-agent"))

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
			"www.gpu-monkey.com",
			"gpu-monkey.com",
			"search.jd.com",
			"jd.com",
			"www.techpowerup.com",
			"techpowerup.com",
		),
		colly.AllowURLRevisit(),
	)

	// hkCollector := collector.Clone()

	GPUData := getGPUSpecData(link, collector)
	GPUData.Name = name
	return GPUData
}

func GetGPUData(specSpec GPUSpecTempStruct, enLink string, cnLink string, hkLink string) GPUType {
	fakeChrome := req.C().ImpersonateChrome().SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36").SetTLSFingerprintChrome()

	fmt.Println(fakeChrome.Headers.Get("user-agent"))

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
			"www.gpu-monkey.com",
			"gpu-monkey.com",
			"search.jd.com",
			"jd.com",
			"www.techpowerup.com",
			"techpowerup.com",
		),
		colly.AllowURLRevisit(),
	)

	// usCollector := collector.Clone()
	cnCollector := collector.Clone()
	// hkCollector := collector.Clone()

	GPUData := GPUType{
		Name:       specSpec.Name,
		Brand:      specSpec.Brand,
		MemorySize: specSpec.MemorySize,
		MemoryType: specSpec.MemoryType,
		MemoryBus:  specSpec.MemoryBus,
		Clock:      specSpec.Clock,
		Power:      specSpec.Power,
		Length:     specSpec.Length,
		Slot:       specSpec.Slot,
		Width:      specSpec.Width,
		PriceUS:    0,
		PriceHK:    0,
		PriceCN:    0,
		Img:        specSpec.Name,
	}
	// GPUData.PriceUS, GPUData.Img = getGPUUSPrice(enLink, usCollector)
	GPUData.PriceCN = getGPUCNPrice(cnLink, cnCollector)
	// GPUData.PriceHK = getGPUHKPrice(hkLink, hkCollector)

	return GPUData
}

func getGPUSpecData(link string, collector *colly.Collector) GPUSpecTempStruct {
	name := ""
	brand := ""
	memorySize := ""
	memoryType := ""
	memoryBus := ""
	clock := ""
	tdp := 0
	length := 0
	slot := ""
	width := 0
	// singleCoreScore := 0
	// muitiCoreScore := 0

	collectorErrorHandle(collector, link)

	collector.OnHTML(".contnt", func(element *colly.HTMLElement) {

		element.ForEach(".sectioncontainer .details .clearfix", func(i int, item *colly.HTMLElement) {
			fmt.Println(item.ChildText("dt"))
			switch item.ChildText("dt") {
			case "Based on:":
				brand = item.ChildText("dd")
			case "Memory Size":
				memorySize = item.ChildText("dd")
			case "Memory Type":
				memoryType = item.ChildText("dd")
			case "Memory Bus":
				memoryBus = item.ChildText("dd")
			case "Boost Clock":
				clock = item.ChildText("dd")
			case "TDP":
				tdp = extractNumberFromString(item.ChildText("dd"))
			case "Length":
				length = extractNumberFromString(item.ChildText("dd"))
			case "Slot Width":
				slot = item.ChildText("dd")
			case "Width":
				width = extractNumberFromString(item.ChildText("dd"))
			}
		})

		name = element.ChildText(".card-head .title-h1")
		/*
			fmt.Println("record logic!!")
			fmt.Println(brand)
			fmt.Println(cores)
			fmt.Println(thread)
			fmt.Println(socket)
			fmt.Println(singleCoreScore)
			fmt.Println(muitiCoreScore)
			fmt.Println(gpu)
			fmt.Println(tdp)
		*/
	})

	collector.Visit("https://www.techpowerup.com/gpu-specs/geforce-rtx-3060-12-gb.c3682")

	fmt.Println(GPUType{
		Name:       name,
		Brand:      brand,
		MemorySize: memorySize,
		MemoryType: memoryType,
		MemoryBus:  memoryBus,
		Clock:      clock,
		Power:      tdp,
		Length:     length,
		Slot:       slot,
		Width:      width,
	})

	return GPUSpecTempStruct{
		Name:       name,
		Brand:      brand,
		MemorySize: memorySize,
		MemoryType: memoryType,
		MemoryBus:  memoryBus,
		Clock:      clock,
		Power:      tdp,
		Length:     length,
		Slot:       slot,
		Width:      width,
	}
}

func getGPUUSPrice(link string, collector *colly.Collector) (float64, string) {
	imgLink := ""
	price := 0.0

	collectorErrorHandle(collector, link)
	fmt.Println(collector.AllowedDomains)

	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")

		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current")), 64); err == nil {
			price = s
			//fmt.Println(price)
		}
	})

	collector.Visit(link)
	return price, imgLink
}

func getGPUHKPrice(link string, collector *colly.Collector) float64 {
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

func getGPUCNPrice(link string, collector *colly.Collector) float64 {
	price := 0.0

	collectorErrorHandle(collector, link)

	collector.OnHTML(".product-mallSales", func(element *colly.HTMLElement) {
		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText("em.price")), 64); err == nil {
			price = s
			fmt.Println(price)
		} else {
			fmt.Println(err)
		}
	})

	collector.Visit(link)
	return price
}
