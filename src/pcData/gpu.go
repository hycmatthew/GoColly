package pcData

import (
	"fmt"
	"math"
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
	Series     string
	Generation string
	MemorySize string
	MemoryType string
	MemoryBus  string
	Clock      int
	Score      int
	Power      int
	Length     int
	Slot       string
	Width      int
	PriceUS    float64
	PriceHK    float64
	PriceCN    float64
	Img        string
}

func GetGPUSpec(name string, link string, score int) GPUSpecTempStruct {
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
	GPUData.Series = name
	GPUData.Score = score
	return GPUData
}

func GetGPUData(specList []GPUSpecTempStruct, name string, brand string, enLink string, cnLink string, hkLink string) GPUType {
	fakeChrome := req.C().ImpersonateChrome().SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36").SetTLSFingerprintChrome()

	fmt.Println(fakeChrome.Headers.Get("user-agent"))

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"nanoreview.net",
			"www.newegg.com",
			"newegg.com",
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
		),
		colly.AllowURLRevisit(),
	)

	usCollector := collector.Clone()
	cnCollector := collector.Clone()
	// hkCollector := collector.Clone()

	priceUs, gpuImg, specUpdate := getGPUUSPrice(enLink, usCollector)
	priceCn, gpuName := getGPUCNPrice(cnLink, cnCollector)
	// GPUData.PriceHK = getGPUHKPrice(hkLink, hkCollector)
	specData := findGPUSpecLogic(specList, gpuName)
	updatedData := searchSubDataByName(name, brand, specData.ProductSpec)

	clockLogic := updatedData.BoostClock
	if specUpdate.BoostClock != 0 {
		clockLogic = specUpdate.BoostClock
	}

	tdpLogic := updatedData.TDP
	if specUpdate.TDP > 0 {
		tdpLogic = specUpdate.TDP
	}

	lengthLogic := updatedData.Length
	if specUpdate.Length > 0 {
		lengthLogic = specUpdate.Length
	}

	newScore := newScoreLogic(clockLogic, specData.Clock, specData.Score)

	GPUData := GPUType{
		Name:       name,
		Brand:      brand,
		Series:     specData.Series,
		Generation: specData.Generation,
		MemorySize: specData.MemorySize,
		MemoryType: specData.MemoryType,
		MemoryBus:  specData.MemoryBus,
		Score:      newScore,
		Clock:      clockLogic,
		Power:      tdpLogic,
		Length:     lengthLogic,
		Slot:       specData.Slot,
		Width:      specData.Width,
		PriceUS:    priceUs,
		PriceHK:    0,
		PriceCN:    priceCn,
		Img:        gpuImg,
	}

	return GPUData
}

func getGPUSpecData(link string, collector *colly.Collector) GPUSpecTempStruct {
	name := ""
	generation := ""
	memorySize := ""
	memoryType := ""
	memoryBus := ""
	clock := 0
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
				clock = extractNumberFromString(item.ChildText("dd"))
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

		element.ForEach(".details.customboards tbody tr", func(i int, item *colly.HTMLElement) {
			splitData := strings.Split(item.ChildText("td:nth-child(5)"), ",")
			tempLength := length
			tempTdp := tdp
			tempSlots := slot

			for i := range splitData {
				if strings.Contains(splitData[i], "mm") {
					tempLength = extractNumberFromString(strings.Split(splitData[i], "mm")[0])
				}
				if strings.Contains(splitData[i], "W") {
					tempTdp = extractNumberFromString(strings.Split(splitData[i], "W")[0])
				}
				if strings.Contains(splitData[i], "slot") {
					tempSlots = splitData[i]
				}
			}

			subData := GPUSpecSubData{
				ProductName: item.ChildText("td:nth-child(1)"),
				BoostClock:  extractNumberFromString(item.ChildText("td:nth-child(3)")),
				Length:      tempLength,
				Slots:       tempSlots,
				TDP:         tempTdp,
			}
			subDataList = append(subDataList, subData)
		})
	})

	collector.Visit(link)

	return GPUSpecTempStruct{
		Series:      name,
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

func getGPUUSPrice(link string, collector *colly.Collector) (float64, string, GPUSpecSubData) {
	imgLink := ""
	price := 0.0
	specSubData := GPUSpecSubData{}

	collectorErrorHandle(collector, link)
	fmt.Println(collector.AllowedDomains)

	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")

		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current")), 64); err == nil {
			price = s
			//fmt.Println(price)
		}

		element.ForEach(".product-details .tab-panes tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("th") {
			case "Boost Clock":
				specSubData.BoostClock = extractNumberFromString(item.ChildText("dd"))
			case "Thermal Design Power":
				specSubData.TDP = extractNumberFromString(item.ChildText("dd"))
			case "Max GPU Length":
				specSubData.Length = extractNumberFromString(item.ChildText("dd"))
			}
		})
	})

	collector.Visit(link)
	return price, imgLink, specSubData
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
		gpuName = extractGPUStringFromString(element.ChildText(".baseParam i:nth-child(2)"))
		fmt.Println(gpuName)
	})

	collector.Visit(link)
	return price, gpuName
}

func findGPUSpecLogic(specList []GPUSpecTempStruct, matchName string) GPUSpecTempStruct {
	for i := range specList {
		if specList[i].Series == matchName {
			return specList[i]
		}
	}
	return GPUSpecTempStruct{}
}

func newScoreLogic(boostClock int, baseClock int, score int) int {
	clockFactor := float64(boostClock) / float64(baseClock)
	updatedScore := float64(score) * (clockFactor * 0.6)
	return int(math.Floor(updatedScore))
}
