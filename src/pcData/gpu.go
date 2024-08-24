package pcData

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type GPURecord struct {
	Name   string
	Brand  string
	LinkCN string
	LinkHK string
	LinkUS string
}

type GPUType struct {
	Name       string
	Brand      string
	Generation string
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

func GetGPUData(specList []GPUSpecTempStruct, brand string, enLink string, cnLink string, hkLink string) GPUType {
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

	// GPUData.PriceUS, GPUData.Img = getGPUUSPrice(enLink, usCollector)
	priceCn, gpuName := getGPUCNPrice(cnLink, cnCollector)
	// GPUData.PriceHK = getGPUHKPrice(hkLink, hkCollector)
	specData := findGPUSpecLogic(specList, gpuName)

	GPUData := GPUType{
		Name:       specData.Name,
		Brand:      brand,
		MemorySize: specData.MemorySize,
		MemoryType: specData.MemoryType,
		MemoryBus:  specData.MemoryBus,
		Clock:      specData.Clock,
		Power:      specData.Power,
		Length:     specData.Length,
		Slot:       specData.Slot,
		Width:      specData.Width,
		PriceUS:    0,
		PriceHK:    0,
		PriceCN:    priceCn,
		Img:        specData.Name,
	}

	return GPUData
}

func getGPUSpecData(link string, collector *colly.Collector) GPUSpecTempStruct {
	name := ""
	generation := ""
	memorySize := ""
	memoryType := ""
	memoryBus := ""
	clock := ""
	tdp := 0
	length := 0
	slot := ""
	width := 0
	var subDataList []GPUSpecSubData

	collectorErrorHandle(collector, link)

	collector.OnHTML(".contnt", func(element *colly.HTMLElement) {

		element.ForEach(".sectioncontainer .details .clearfix", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("dt") {
			case "Generation":
				generation = item.ChildText("dd")
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

		element.ForEach(".details.customboards tbody tr", func(i int, item *colly.HTMLElement) {
			splitData := strings.Split(item.ChildText("td:nth-child(5)"), ",")
			tempLength := ""
			tempTdp := ""
			tempSlots := ""

			for i := range splitData {
				if strings.Contains(splitData[i], "mm") {
					tempLength = strings.Split(splitData[i], "mm")[0]
				}
				if strings.Contains(splitData[i], "W") {
					tempTdp = strings.Split(splitData[i], "W")[0]
				}
				if strings.Contains(splitData[i], "slot") {
					tempSlots = splitData[i]
				}
			}

			subData := GPUSpecSubData{
				ProductName: item.ChildText("td:nth-child(1)"),
				BoostClock:  item.ChildText("td:nth-child(3)"),
				Length:      tempLength,
				Slots:       tempSlots,
				TDP:         tempTdp,
			}

			fmt.Println(subData)

			subDataList = append(subDataList, subData)
		})
	})

	collector.Visit(link)

	return GPUSpecTempStruct{
		Name:        name,
		Generation:  generation,
		MemorySize:  memorySize,
		MemoryType:  memoryType,
		MemoryBus:   memoryBus,
		Clock:       clock,
		Power:       tdp,
		Length:      length,
		Slot:        slot,
		Width:       width,
		ProductSpec: subDataList,
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

func getGPUCNPrice(link string, collector *colly.Collector) (float64, string) {
	price := 0.0
	gpuName := ""

	collectorErrorHandle(collector, link)

	collector.OnHTML(".product-detail-main", func(element *colly.HTMLElement) {
		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText(".product-mallSales em.price")), 64); err == nil {
			price = s
			fmt.Println(price)
		} else {
			fmt.Println(err)
		}
		gpuName = extractGPUStringFromString(element.ChildText(".baseParam i.u-longTxt:first-child"))
		fmt.Println(gpuName)
	})

	collector.Visit(link)
	return price, gpuName
}

func findGPUSpecLogic(specList []GPUSpecTempStruct, matchName string) GPUSpecTempStruct {
	for i := range specList {
		if specList[i].Name == matchName {
			return specList[i]
		}
	}
	return GPUSpecTempStruct{}
}
