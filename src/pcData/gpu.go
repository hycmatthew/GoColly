package pcData

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type GPUScoreData struct {
	Name      string
	Benchmark string
	DataLink  string
}

type GPURecordData struct {
	Brand   string
	Name    string
	PriceCN string
	SpecCN  string
	LinkCN  string
	LinkUS  string
	LinkHK  string
}

type GPUSpec struct {
	Manufacturer string
	Series       string
	Generation   string
	MemorySize   int
	MemoryType   string
	MemoryBus    string
	BoostClock   int
	Benchmark    int
	Power        int
	Length       int
	Slot         string
	Width        int
	ProductSpec  []GPUSpecSubData
}

type GPUSpecSubData struct {
	ProductName string
	BoostClock  int
	Length      int
	Slots       string
	TDP         int
}

type GPUType struct {
	Id           string
	Name         string
	Brand        string
	Manufacturer string
	Series       string
	Generation   string
	MemorySize   int
	MemoryType   string
	MemoryBus    string
	BoostClock   int
	OcClock      int
	Benchmark    int
	Power        int
	Length       int
	Slot         string
	LinkCN       string
	LinkUS       string
	LinkHK       string
	PriceUS      string
	PriceHK      string
	PriceCN      string
	Img          string
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
	// scoreCollector := collector.Clone()

	GPUSpec := getGPUSpecData(record.DataLink, collector)
	GPUSpec.Series = record.Name
	GPUSpec.Benchmark = extractNumberFromString(record.Benchmark)
	return GPUSpec
}

func GetGPUData(specList []GPUSpec, record GPURecordData) (GPUType, bool) {
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
			"www.colorful.cn",
			"colorful.cn",
		),
		colly.AllowURLRevisit(),
	)

	usCollector := collector.Clone()
	cnCollector := collector.Clone()
	// hkCollector := collector.Clone()

	specData := GPUSpec{}
	ocClock := 0
	priceUs, gpuImg, specUpdate := "", "", GPUSpecSubData{}
	isValid := true

	if record.LinkUS != "" {
		priceUs, gpuImg, specUpdate = getGPUUSPrice(record.LinkUS, usCollector)
		if priceUs == "" {
			isValid = false
		}
	}
	priceCn, tempSeries := getGPUCNPrice(record.LinkCN, cnCollector)
	seriesName := updateSeriesName(record.Name, tempSeries)
	// GPUData.PriceHK = getGPUHKPrice(hkLink, hkCollector)

	if priceCn == "" {
		isValid = false
	}

	switch record.Brand {
	case "colorful":
		specData, ocClock, gpuImg = getSpecFromColorful(record.SpecCN)
		specData.Benchmark = findGPUScoreLogic(specList, seriesName)
	default:
		specData = findGPUSpecLogic(specList, seriesName)
		ocClock = specData.BoostClock

		searchSubData := searchSubDataByName(record.Name, record.Brand, specData.ProductSpec)
		if searchSubData.BoostClock != 0 {
			specUpdate = searchSubData
		}
		specData.BoostClock = specUpdate.BoostClock
		specData.Length = specUpdate.Length
		specData.Power = specUpdate.TDP
		specData.Slot = specUpdate.Slots
	}

	newBenchmark := int(newScoreLogic(ocClock, specData.BoostClock, specData.Benchmark))
	GPUData := GPUType{
		Id:           SetProductId(record.Brand, record.Name),
		Name:         record.Name,
		Brand:        record.Brand,
		Manufacturer: specData.Manufacturer,
		Series:       specData.Series,
		Generation:   specData.Generation,
		MemorySize:   specData.MemorySize,
		MemoryType:   specData.MemoryType,
		MemoryBus:    specData.MemoryBus,
		Benchmark:    newBenchmark,
		BoostClock:   specData.BoostClock,
		OcClock:      ocClock,
		Power:        specData.Power,
		Length:       specData.Length,
		Slot:         specData.Slot,
		LinkUS:       record.LinkUS,
		LinkHK:       record.LinkHK,
		LinkCN:       record.LinkCN,
		PriceUS:      priceUs,
		PriceHK:      "",
		PriceCN:      priceCn,
		Img:          gpuImg,
	}

	return GPUData, isValid
}

func getGPUSpecData(link string, collector *colly.Collector) GPUSpec {
	specData := GPUSpec{}
	var subDataList []GPUSpecSubData

	collectorErrorHandle(collector, link)

	collector.OnHTML(".contnt", func(element *colly.HTMLElement) {
		tempName := strings.ToUpper(element.ChildText(".gpudb-name"))
		if strings.Contains(tempName, "NVIDIA") {
			specData.Manufacturer = "NVIDIA"
		} else {
			specData.Manufacturer = "AMD"
		}

		element.ForEach(".sectioncontainer .details .clearfix", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("dt") {
			case "Generation":
				tempString := item.ChildText("dd")
				if strings.Contains(tempString, "(") {
					genString := strings.Split(item.ChildText("dd"), "(")
					specData.Generation = strings.ReplaceAll(genString[1], ")", "")
				} else {
					specData.Generation = tempString
				}
			case "Memory Size":
				specData.MemorySize = extractNumberFromString(item.ChildText("dd"))
			case "Memory Type":
				specData.MemoryType = item.ChildText("dd")
			case "Memory Bus":
				specData.MemoryBus = item.ChildText("dd")
			case "Boost Clock":
				specData.BoostClock = extractNumberFromString(item.ChildText("dd"))
			case "TDP":
				specData.Power = extractNumberFromString(item.ChildText("dd"))
			case "Length":
				specData.Length = extractNumberFromString(item.ChildText("dd"))
			case "Slot Width":
				specData.Slot = item.ChildText("dd")
			case "Width":
				specData.Width = extractNumberFromString(item.ChildText("dd"))
			}
		})

		element.ForEach(".details.customboards tbody tr", func(i int, item *colly.HTMLElement) {
			splitData := strings.Split(item.ChildText("td:nth-child(5)"), ",")
			tempLength := specData.Length
			tempTdp := specData.Power
			tempSlots := specData.Slot

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
		specData.ProductSpec = subDataList
	})

	collector.Visit(link)

	return specData
}

// https://nanoreview.net/en/gpu/radeon-rx-6600-xt
/*
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
*/

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

func findGPUSpecLogic(specList []GPUSpec, matchSeries string) GPUSpec {
	var tempData GPUSpec
	upperSeries := strings.ToUpper(matchSeries)
	for _, item := range specList {
		if strings.ToUpper(item.Series) == upperSeries {
			tempData = item
			break
		}
	}
	return tempData
}

func findGPUScoreLogic(specList []GPUSpec, matchSeries string) int {
	var tempData GPUSpec
	upperSeries := strings.ToUpper(matchSeries)
	for _, item := range specList {
		if strings.ToUpper(item.Series) == upperSeries {
			tempData = item
			break
		}
	}
	return tempData.Benchmark
}

func newScoreLogic(ocClock int, boostClock int, score int) float64 {
	resScore := float64(score)
	if ocClock != boostClock {
		clockFactor := float64(ocClock) / float64(boostClock)
		updatedFactor := ((clockFactor - 1) * 0.6) + 1
		resScore = resScore * updatedFactor
	}
	return math.Floor(resScore)
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
		{"AERO", "S"},
	}

	resSeries := msiSeries

	switch brand {
	case "asus":
		resSeries = asusSeries
	case "colorful":
		resSeries = colorfulSeries
	case "galaxy":
		resSeries = galaxySeries
	case "gigabyte":
		resSeries = gigabyteSeries
	default:
		resSeries = msiSeries
	}

	for i := range resSeries {
		for j := range resSeries[i] {
			resSeries[i][j] = strings.ToUpper(resSeries[i][j])
		}
	}
	return resSeries
}

func searchSubDataByName(name string, brand string, subDataList []GPUSpecSubData) GPUSpecSubData {
	brandStr := strings.ToLower(brand)
	seriesList := getBrandSeries(brandStr)

	updatedName := strings.ToUpper(strings.ReplaceAll(name, "-", " "))
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
	fmt.Println(name, "is OC : ", isOC)
	for _, item := range seriesList {
		if isSubset(item, nameList) {
			matchedseries = item
		}
	}

	var out GPUSpecSubData
	tempSubdDataList := filterByBrand(brandStr, subDataList)
	for _, item := range tempSubdDataList {
		upperName := strings.ToUpper(item.ProductName)
		subDataNameList := strings.Split(upperName, " ")
		subOC := strings.Contains(upperName, " OC")

		if isSubset(matchedseries, subDataNameList) && isOC == subOC {
			fmt.Println(name, " Matched sub data : ", item.ProductName)
			out = item
			break
		}
	}
	if out.BoostClock == 0 {
		fmt.Println(updatedName, " - cant find sub data - ", matchedseries)
	}
	return out
}

type ColorfulSpecData struct {
	ArticleId   string `json:"article_id"`
	AttributeId string `json:"attribute_id"`
	Content     string `json:"content"`
	CreateTime  string `json:"create_time"`
	Id          string `json:"id"`
	Mid         string `json:"mid"`
	Title       string `json:"title"`
	Typeid      string `json:"typeid"`
}

func getSpecFromColorful(link string) (GPUSpec, int, string) {
	tempLink := strings.Split(link, "?")[1]
	response, err := http.Get("https://www.colorful.cn/Home/GetAttrbuteValue?" + tempLink)
	if err != nil {
		fmt.Println("无法发起请求:", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("无法读取响应体:", err)
	}

	var jsonObj []ColorfulSpecData
	json.Unmarshal(body, &jsonObj)

	specData := GPUSpec{}
	ocClock := 0
	imgLink := ""

	for _, element := range jsonObj {
		if element.Title == "芯片系列" {
			tempSeries := strings.Split(element.Content, "; ")
			specData.Series = tempSeries[len(tempSeries)-1]
		}
		if element.Title == "显存容量" {
			specData.MemorySize = extractNumberFromString(element.Content)
		}
		if element.Title == "显存类型" {
			specData.MemoryType = element.Content
		}
		if element.Title == "显存位宽" {
			specData.MemoryBus = element.Content
		}
		if element.Title == "TDP功耗" {
			specData.Power = extractNumberFromString(element.Content)
		}
		if element.Title == "基础频率" {
			tempClock := strings.Split(element.Content, "Boost")
			specData.BoostClock = extractNumberFromString(tempClock[1])

			if ocClock == 0 {
				ocClock = specData.BoostClock
			}
		}
		if element.Title == "一键OC核心频率" {
			tempClock := strings.Split(element.Content, "Boost")
			ocClock = extractNumberFromString(tempClock[1])
		}
		if element.Title == "产品尺寸" {
			specData.Length = extractNumberFromString(element.Content)
		}
		if element.Title == "显卡类型" {
			specData.Slot = slotTranslation(element.Content)
		}
		if element.Title == "产品图片" {
			imgLink = "https://www.colorful.cn/" + element.Content
		}
	}

	return specData, ocClock, imgLink
}

func updateSeriesName(gpuName string, series string) string {
	res := series
	if strings.Contains(series, "Radeon") {
		updatedName := strings.Split(series, "Radeon")
		res = strings.TrimSpace(updatedName[len(updatedName)-1])
	}

	if strings.Contains(series, "GeForce") {
		updatedName := strings.Split(series, "GeForce")
		res = strings.TrimSpace(updatedName[len(updatedName)-1])
	}

	if res == "RTX 4060 Ti" {
		if strings.Contains(strings.ToUpper(gpuName), "8G") {
			res = "RTX 4060 Ti 8GB"
		}
		if strings.Contains(strings.ToUpper(gpuName), "16G") {
			res = "RTX 4060 Ti 16GB"
		}
	}
	return res
}

func slotTranslation(str string) string {
	slotTranslation := map[string]string{
		"双槽":  "2 Slots",
		"超双槽": "2.5 Slots",
		"三槽":  "3 Slots",
		"超三槽": "3.5 Slots",
	}

	res, valid := slotTranslation[str]

	if valid {
		return res
	} else {
		return str
	}
}

/*
func getWidthFromDimension(dimension string) int {
	numList := strings.Split(dimension, "*")
	return extractNumberFromString(numList[0])
}
*/

func CompareGPUDataLogic(cur GPUType, list []GPUType) GPUType {
	newVal := cur
	curTest := cur.Brand + cur.Name
	oldVal := cur
	for _, item := range list {
		testStr := item.Brand + item.Name
		if curTest == testStr {
			oldVal = item
			break
		}
	}

	if newVal.PriceCN == "" {
		newVal.PriceCN = oldVal.PriceCN
	}
	if newVal.PriceUS == "" {
		newVal.PriceUS = oldVal.PriceUS
	}
	if newVal.PriceHK == "" {
		newVal.PriceHK = oldVal.PriceHK
	}
	return newVal
}
