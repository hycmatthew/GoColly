package pcData

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type GPUScoreData struct {
	Name      string
	DataLink  string
	ScoreLink string
}

type GPURecordData struct {
	Brand   string
	Name    string
	PriceCN string
	LinkCN  string
	LinkUS  string
	LinkHK  string
}

type GPUSpec struct {
	Code        string
	Series      string
	Generation  string
	MemorySize  string
	MemoryType  string
	MemoryBus   string
	Clock       int
	TimeSpy     int
	FrameScore  float64
	Power       int
	Length      int
	Slot        string
	Width       int
	ProductSpec []GPUSpecSubData
}

type GPUSpecSubData struct {
	ProductName string
	BoostClock  int
	Length      int
	Slots       string
	TDP         int
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
	TimeSpy    int
	FrameScore float64
	Power      int
	Length     int
	Slot       string
	Width      int
	PriceUS    string
	PriceHK    string
	PriceCN    string
	Img        string
}

func GetGPUSpec(record GPUScoreData) GPUSpec {
	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"nanoreview.net",
			"www.newegg.com",
			"newegg.com",
			"www.techpowerup.com",
			"techpowerup.com",
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})
	scoreCollector := collector.Clone()

	GPUSpec := getGPUSpecData(record.DataLink, collector)
	GPUSpec.TimeSpy, GPUSpec.FrameScore = getGPUScoreData(record.ScoreLink, scoreCollector)
	GPUSpec.Code = record.Name
	return GPUSpec
}

func GetGPUData(specList []GPUSpec, record GPURecordData) GPUType {
	fakeChrome := req.DefaultClient().ImpersonateChrome()
	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"nanoreview.net",
			"www.newegg.com",
			"newegg.com",
			"www.price.com.hk",
			"price.com.hk",
			"detail.zol.com.cn",
			"zol.com.cn",
			"product.pconline.com.cn",
			"pconline.com.cn",
		),
		colly.AllowURLRevisit(),
	)

	usCollector := collector.Clone()
	cnCollector := collector.Clone()
	// hkCollector := collector.Clone()

	priceUs, gpuImg, specUpdate := getGPUUSPrice(record.LinkUS, usCollector)
	priceCn, gpuName := getGPUCNPrice(record.LinkCN, cnCollector)
	// GPUData.PriceHK = getGPUHKPrice(hkLink, hkCollector)
	specData := findGPUSpecLogic(specList, gpuName)
	updatedData := searchSubDataByName(record.Name, record.Brand, specData.ProductSpec)

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

	newTimeSpy := int(newScoreLogic(clockLogic, specData.Clock, float64(specData.TimeSpy)))
	newFrameScore := newScoreLogic(clockLogic, specData.Clock, specData.FrameScore)

	GPUData := GPUType{
		Name:       record.Name,
		Brand:      record.Brand,
		Series:     specData.Series,
		Generation: specData.Generation,
		MemorySize: specData.MemorySize,
		MemoryType: specData.MemoryType,
		MemoryBus:  specData.MemoryBus,
		TimeSpy:    newTimeSpy,
		FrameScore: newFrameScore,
		Clock:      clockLogic,
		Power:      tdpLogic,
		Length:     lengthLogic,
		Slot:       specData.Slot,
		Width:      specData.Width,
		PriceUS:    priceUs,
		PriceHK:    "",
		PriceCN:    priceCn,
		Img:        gpuImg,
	}

	return GPUData
}

func getGPUSpecData(link string, collector *colly.Collector) GPUSpec {
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

	return GPUSpec{
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

func getGPUScoreData(link string, collector *colly.Collector) (int, float64) {
	timespy := 0
	framescore := 0.0
	collectorErrorHandle(collector, link)

	collector.OnHTML("#the-app", func(element *colly.HTMLElement) {
		element.ForEach(".two-columns-item .score-bar", func(i int, item *colly.HTMLElement) {
			switch item.ChildText(".score-bar-name") {
			case "Time Spy Score":
				timespy = extractNumberFromString(item.ChildText(".score-bar-result-number"))
			case "Aztec Ruins High Tier 4K (FPS)":
				framescore, _ = strconv.ParseFloat(extractFloatStringFromString(item.ChildText(".score-bar-result-number")), 64)
			}
		})
	})
	collector.Visit(link)
	return timespy, framescore
}

func getGPUUSPrice(link string, collector *colly.Collector) (string, string, GPUSpecSubData) {
	imgLink := ""
	price := ""
	specSubData := GPUSpecSubData{}

	collectorErrorHandle(collector, link)

	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")

		price = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current"))
		element.ForEach(".product-details .tab-panes tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("th") {
			case "Boost Clock":
				specSubData.BoostClock = extractNumberFromString(item.ChildText("td"))
			case "Thermal Design Power":
				specSubData.TDP = extractNumberFromString(item.ChildText("td"))
			case "Max GPU Length":
				specSubData.Length = extractNumberFromString(item.ChildText("td"))
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
			// fmt.Println(extractFloatStringFromString(element.ChildText("span")))
			if price == 0.0 {
				if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText("span")), 64); err == nil {
					price = s
				} else {
					fmt.Println(err)
				}
			}
		})
	})

	collector.Visit(link)
	return price
}

func getGPUCNPrice(link string, collector *colly.Collector) (string, string) {
	price := ""
	gpuName := ""

	collectorErrorHandle(collector, link)

	collector.OnHTML(".product-detail-main", func(element *colly.HTMLElement) {
		price = extractFloatStringFromString(element.ChildText(".product-mallSales em.price"))
		gpuName = extractGPUStringFromString(element.ChildText(".baseParam i:nth-child(2)"))
	})

	collector.Visit(link)
	return price, gpuName
}

func findGPUSpecLogic(specList []GPUSpec, matchName string) GPUSpec {
	var tempData GPUSpec
	upperNameList := strings.Split(strings.ToUpper(matchName), " ")
	seriesLength := 0
	for i := range specList {
		seriesList := strings.Split(specList[i].Series, " ")

		if isSubset(seriesList, upperNameList) && (len(seriesList) > seriesLength) {
			tempData = specList[i]
			seriesLength = len(seriesList)
		}
	}
	return tempData
}

func newScoreLogic(boostClock int, baseClock int, score float64) float64 {
	clockFactor := float64(boostClock) / float64(baseClock)
	updatedScore := score * (clockFactor * 0.6)
	return math.Floor(updatedScore)
}

func filterByBrand(brand string, in []GPUSpecSubData) []GPUSpecSubData {
	var out []GPUSpecSubData
	for _, item := range in {
		brandStr := strings.Split(item.ProductName, " ")[0]
		if strings.ToLower(brandStr) == brand {
			out = append(out, item)
		}
	}
	return out
}

func getBrandSeries(brand string) [][]string {
	asusSeries := [][]string{
		{"DUAL", "V2"},
		{"DUAL", "WHITE"},
		{"MEGALODON"},
		{"PRIME"},
		{"STRIX"},
		{"TUF"},
	}
	colorfulSeries := [][]string{
		{"iGame", "Advanced"},
		{"iGame", "Ultra", "DUO"},
		{"iGame", "Ultra"},
		{"Tomahawk", "Deluxe"},
		{"Tomahawk", "DUO"},
	}
	galaxySeries := [][]string{
		{"Click"},
		{"BOOMSTAR"},
		{"EX"},
		{"EX", "White"},
		{"METALTOP"},
	}

	gigabyteSeries := [][]string{
		{"AORUS", "ELITE"},
		{"AERO"},
		{"EAGLE"},
		{"GAMING"},
		{"WINDFORCE"},
	}

	msiSeries := [][]string{
		{"GAMING"},
		{"GAMING", "TRIO"},
		{"GAMING", "X"},
		{"GAMING", "X", "TRIO"},
		{"VENTUS", "2X"},
		{"VENTUS", "3X"},
		{"GAMING", "X", "SLIM"},
	}

	switch brand {
	case "asus":
		return asusSeries
	case "colorful":
		return colorfulSeries
	case "galaxy":
		return galaxySeries
	case "gigabyte":
		return gigabyteSeries
	default:
		return msiSeries
	}

}

func searchSubDataByName(name string, brand string, subDataList []GPUSpecSubData) GPUSpecSubData {
	brandStr := strings.ToLower(brand)
	seriesList := getBrandSeries(brandStr)
	for i := range seriesList {
		for j := range seriesList[i] {
			seriesList[i][j] = strings.ToUpper(seriesList[i][j])
		}
	}

	updatedName := strings.ToUpper(strings.Replace(name, "-", " ", -1))
	nameList := strings.Split(updatedName, " ")
	var matchedseries []string
	isOC := false
	for _, item := range nameList {
		if brandStr == "asus" {
			first := item[0:1]
			last := item[len(item)-1:]

			if first == "O" && last == "G" {
				isOC = true
			}
		}
		if item == "OC" {
			isOC = true
		}
	}

	for i := range seriesList {
		if isSubset(seriesList[i], nameList) {
			matchedseries = seriesList[i]
		}
	}
	var out GPUSpecSubData
	tempSubdDataList := filterByBrand(brandStr, subDataList)
	for i := range tempSubdDataList {
		upperName := strings.ToUpper(tempSubdDataList[i].ProductName)
		subDataNameList := strings.Split(upperName, " ")
		subOC := strings.Contains(upperName, " OC")

		if isSubset(matchedseries, subDataNameList) && isOC == subOC {
			fmt.Println("Matched sub data - ", tempSubdDataList[i])
			out = tempSubdDataList[i]
		}
	}
	if out.BoostClock == 0 {
		fmt.Println(updatedName, " - cant find sub data - ", matchedseries)
	}
	return out
}
