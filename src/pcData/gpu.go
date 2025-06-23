package pcData

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

type GPUScoreData struct {
	Name      string
	Benchmark string
	DataLink  string
}

type GPUScore struct {
	Manufacturer string
	Chipset      string
	Generation   string
	MemorySize   int
	MemoryType   string
	MemoryBus    string
	ClockRate    int
	BoostClock   int
	Benchmark    int
	Power        int
	Length       int
	Slot         string
	Width        int
	ProductSpec  []GPUScoreSubData
}

type GPUScoreSubData struct {
	ProductName string
	BoostClock  int
	Length      int
	Slots       string
	TDP         int
}

type GPUSpec struct {
	Code         string
	Name         string
	Brand        string
	ReleaseDate  string
	Manufacturer string
	Chipset      string
	Series       string
	MemorySize   int
	MemoryType   string
	MemoryBus    int
	ClockRate    int
	BoostClock   int
	Benchmark    int
	Power        int
	Length       int
	Slot         string
	Prices       []PriceType
	Img          string
}

type GPUType struct {
	Id           string
	Name         string
	NameCN       string
	Brand        string
	Manufacturer string
	Chipset      string
	Series       string
	MemorySize   int
	MemoryType   string
	MemoryBus    int
	ClockRate    int
	BoostClock   int
	Benchmark    int
	Power        int
	Length       int
	Slot         string
	Prices       []PriceType
	Img          string
}

var gpuScoreList []GPUScore
var chipsetMap map[string]string

func init() {
	scoreFile, _ := os.Open("tmp/spec/gpuScoreSpec.json")
	byteValue, _ := io.ReadAll(scoreFile)
	json.Unmarshal([]byte(byteValue), &gpuScoreList)

	chipsetMap = make(map[string]string)
	for _, item := range gpuScoreList {
		key := strings.ReplaceAll(strings.ToUpper(item.Chipset), " ", "")
		chipsetMap[key] = item.Chipset
	}
	fmt.Println(chipsetMap)
}

func GetGPUScoreSpec(record GPUScoreData) GPUScore {
	GPUSpec := fetchGPUScoreData(record.DataLink, CreateCollector())
	GPUSpec.Chipset = record.Name
	GPUSpec.Benchmark = extractNumberFromString(record.Benchmark)
	return GPUSpec
}

func UpdateGPUBenchmarks(data GPUType) GPUType {
	matchedScore := findGPUScoreDataLogic(gpuScoreList, data.Chipset)
	data.Benchmark = newScoreLogic(matchedScore.BoostClock, data.BoostClock, matchedScore.Benchmark)
	fmt.Printf("UpdateGPUBenchmarks Chipset: %s, matchedScore.Benchmark: %d, data.Benchmark: %d\n", data.Chipset, matchedScore.Benchmark, data.Benchmark)
	return data
}

func GetGPUSpec(record LinkRecord) GPUSpec {
	gpuData := GPUSpec{}
	switch record.Brand {
	case "colorful":
		gpuData = fetchSpecFromColorful(record.LinkSpec)
	default:
		if strings.Contains(record.LinkCN, "zol") {
			record.LinkCN = getDetailsLinkFromZol(record.LinkCN, CreateCollector())
		}
		if record.LinkSpec != "" {
			gpuData = fetchGPUSpecData(record.LinkSpec, CreateCollector())
		}
	}

	gpuData.Brand = record.Brand
	gpuData.Code = record.Name
	if gpuData.Name == "" {
		gpuData.Name = record.Name
	}
	gpuData.Name = RemoveBrandsFromName(gpuData.Brand, gpuData.Name)

	// 合并价格链接
	gpuData.Prices = handleSpecPricesLogic(gpuData.Prices, record)
	return gpuData
}

func GetGPUData(spec GPUSpec) (GPUType, bool) {
	isValid := true
	newSpec := spec
	collector := CreateCollector()

	// 遍历所有价格数据进行处理
	for _, price := range newSpec.Prices {
		switch price.Region {
		case "CN":
			/*
				if strings.Contains(price.PriceLink, "zol") {

					tempSpec := getGPUSpecFromZol(price.PriceLink, collector)
					newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(GPUSpec)

					if updatedPrice := getPriceByPlatform(tempSpec.Prices, "CN", Platform_JD); updatedPrice != nil {
						isValid = isValid && checkPriceValid(updatedPrice.Price)
					}

				}
			*/
			if strings.Contains(price.PriceLink, "pconline") {
				if price.Price == "" {
					priceCN := fetchGPUCNPrice(price.PriceLink, collector)
					newSpec.Prices = upsertPrice(newSpec.Prices, PriceType{
						Region:    "CN",
						Platform:  Platform_JD,
						Price:     priceCN,
						PriceLink: price.PriceLink,
					})
					isValid = isValid && checkPriceValid(priceCN)
				}
			}
		case "US":
			if strings.Contains(price.PriceLink, "newegg") {
				tempSpec := fetchGPUUSPrice(price.PriceLink, collector)
				// 合并图片数据
				if newSpec.Img == "" && tempSpec.Img != "" {
					newSpec.Img = tempSpec.Img
				}
				// 合并规格数据
				newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(GPUSpec)

				if updatedPrice := getPriceByPlatform(tempSpec.Prices, "US", Platform_Newegg); updatedPrice != nil {
					isValid = isValid && checkPriceValid(updatedPrice.Price)
				}
			}
		}
	}

	// update chipset and get boost clock and benchmark from score data
	newSpec.Chipset = normalizeChipset(newSpec.Chipset, newSpec.MemorySize)
	matchedScore := findGPUScoreDataLogic(gpuScoreList, newSpec.Chipset)
	if newSpec.BoostClock == 0 {
		searchSubData := searchSubDataByName(newSpec.Name, newSpec.Brand, matchedScore.ProductSpec)
		newSpec.BoostClock = searchSubData.BoostClock
		if newSpec.Length == 0 {
			newSpec.Length = searchSubData.Length
		}
		if newSpec.Power == 0 {
			newSpec.Power = searchSubData.TDP
		}
		if newSpec.Slot == "" {
			newSpec.Slot = searchSubData.Slots
		}
	}
	newBenchmark := newScoreLogic(matchedScore.BoostClock, newSpec.BoostClock, matchedScore.Benchmark)

	GPUData := GPUType{
		Id:           SetProductId(spec.Brand, spec.Name),
		Name:         newSpec.Name,
		NameCN:       newSpec.Name,
		Brand:        newSpec.Brand,
		Manufacturer: newSpec.Manufacturer,
		Series:       newSpec.Series,
		Chipset:      newSpec.Chipset,
		MemorySize:   newSpec.MemorySize,
		MemoryType:   newSpec.MemoryType,
		MemoryBus:    newSpec.MemoryBus,
		Benchmark:    newBenchmark,
		ClockRate:    newSpec.ClockRate,
		BoostClock:   newSpec.BoostClock,
		Power:        newSpec.Power,
		Length:       newSpec.Length,
		Slot:         newSpec.Slot,
		Prices:       deduplicatePrices(newSpec.Prices),
		Img:          newSpec.Img,
	}

	return GPUData, isValid
}

func fetchGPUScoreData(link string, collector *colly.Collector) GPUScore {
	specData := GPUScore{}
	var subDataList []GPUScoreSubData

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
			case "Base Clock":
				specData.ClockRate = extractNumberFromString(item.ChildText("dd"))
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

			subData := GPUScoreSubData{
				ProductName: item.ChildText("td:nth-child(1)"),
				BoostClock:  extractNumberFromString(item.ChildText("td:nth-child(3)")),
				Length:      tempLength,
				Slots:       strings.TrimSpace(tempSlots),
				TDP:         tempTdp,
			}
			subDataList = append(subDataList, subData)
		})
		specData.ProductSpec = subDataList
	})

	collector.Visit(link)

	return specData
}

func fetchGPUSpecData(link string, collector *colly.Collector) GPUSpec {
	specData := GPUSpec{}
	collectorErrorHandle(collector, link)

	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = strings.TrimSpace(element.ChildText(".breadcrumb .active"))
		specData.Img = element.ChildAttr(".tns-inner .tns-item img", "src")
		specData.Prices = GetPriceLinkFromPangoly(element)

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Release Date":
				specData.ReleaseDate = strings.TrimSpace(item.ChildText("td span"))
			case "Series":
				specData.Series = strings.TrimSpace(item.ChildText("td:nth-of-type(2)"))
				manufacturerStr := strings.Split(specData.Series, " ")
				specData.Manufacturer = manufacturerStr[0]
			case "GPU Chipset":
				specData.Chipset = strings.TrimSpace(item.ChildText("td:nth-of-type(2)"))
			case "GPU Memory Size":
				specData.MemorySize = extractNumberFromString(item.ChildText("td:nth-of-type(2)"))
			case "GPU Memory Type":
				specData.MemoryType = strings.TrimSpace(item.ChildText("td:nth-of-type(2)"))
			case "GPU Memory Interface":
				specData.MemoryBus = extractNumberFromString(strings.TrimSpace(item.ChildText("td:nth-of-type(2)")))
			case "GPU Clock Rate":
				specData.ClockRate = extractNumberFromString(item.ChildText("td:nth-of-type(2)"))
			case "GPU Boost Clock Rate":
				specData.BoostClock = extractNumberFromString(item.ChildText("td:nth-of-type(2)"))
			case "TDP":
				specData.Power = extractNumberFromString(item.ChildText("td:nth-of-type(2)"))
			case "Length":
				specData.Length = extractNumberFromString(item.ChildText("td:nth-of-type(2)"))
			case "Expansion slots required":
				specData.Slot = strings.TrimSpace(item.ChildText("td:nth-of-type(2)"))
			}
		})
	})
	collector.Visit(link)
	return specData
}

func fetchGPUUSPrice(link string, collector *colly.Collector) GPUSpec {
	specData := GPUSpec{}

	collectorErrorHandle(collector, link)

	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		specData.Prices = upsertPrice(specData.Prices, PriceType{
			Region:    "US",
			Platform:  Platform_Newegg,
			Price:     extractFloatStringFromString(element.ChildText(".row-side .product-buy-box .price-current")),
			PriceLink: link,
		})

		element.ForEach(".product-details .tab-panes tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("th") {
			case "Boost Clock":
				specData.BoostClock = extractNumberFromString(item.ChildText("td"))
			case "Thermal Design Power":
				specData.Power = extractNumberFromString(item.ChildText("td"))
			case "Max GPU Length":
				specData.Length = extractNumberFromString(item.ChildText("td"))
			}
		})
	})

	collector.Visit(link)
	return specData
}

func fetchGPUCNPrice(link string, collector *colly.Collector) string {
	price := ""
	// gpuName := ""

	collectorErrorHandle(collector, link)

	collector.OnHTML(".product-detail-main", func(element *colly.HTMLElement) {
		price = extractFloatStringFromString(element.ChildText(".product-mallSales em.price"))
		// gpuName = extractGPUStringFromString(element.ChildText(".baseParam i:nth-child(2)"))
	})

	collector.Visit(link)
	return price
}

func findGPUScoreDataLogic(specList []GPUScore, matchChipset string) GPUScore {
	var tempData GPUScore
	upperChipset := strings.ToUpper(matchChipset)
	for _, item := range specList {
		if strings.ToUpper(item.Chipset) == upperChipset {
			tempData = item
			break
		}
	}
	return tempData
}

func newScoreLogic(origin int, boost int, score int) int {
	originClock := float64(origin)
	boostClock := float64(boost)
	resScore := float64(score)
	if originClock != boostClock {
		clockFactor := boostClock / originClock
		updatedFactor := ((clockFactor - 1) * 0.6) + 1
		resScore = resScore * updatedFactor
	}
	return int(resScore)
}

func filterByBrand(brand string, in []GPUScoreSubData) []GPUScoreSubData {
	var out []GPUScoreSubData
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

func searchSubDataByName(name string, brand string, subDataList []GPUScoreSubData) GPUScoreSubData {
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

	var out GPUScoreSubData
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

func fetchSpecFromColorful(link string) GPUSpec {
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

	for _, element := range jsonObj {
		if element.Title == "芯片系列" {
			tempSeries := strings.Split(element.Content, "; ")
			specData.Chipset = tempSeries[len(tempSeries)-1]
		}
		if element.Title == "显存容量" {
			specData.MemorySize = extractNumberFromString(element.Content)
		}
		if element.Title == "显存类型" {
			specData.MemoryType = element.Content
		}
		if element.Title == "显存位宽" {
			specData.MemoryBus = extractNumberFromString(element.Content)
		}
		if element.Title == "TDP功耗" {
			specData.Power = extractNumberFromString(element.Content)
		}
		if element.Title == "基础频率" {
			tempClock := strings.Split(element.Content, "Boost")
			specData.ClockRate = extractNumberFromString(tempClock[1])

			if specData.BoostClock == 0 {
				specData.BoostClock = specData.ClockRate
			}
		}
		if element.Title == "一键OC核心频率" {
			tempClock := strings.Split(element.Content, "Boost")
			specData.BoostClock = extractNumberFromString(tempClock[1])
		}
		if element.Title == "产品尺寸" {
			specData.Length = extractNumberFromString(element.Content)
		}
		if element.Title == "显卡类型" {
			specData.Slot = slotTranslation(element.Content)
		}
		if element.Title == "产品图片" {
			specData.Img = "https://www.colorful.cn/" + element.Content
		}
	}
	return specData
}

func normalizeChipset(input string, vram int) string {
	// 移除品牌前缀并标准化格式
	re := regexp.MustCompile(`(?i)^\s*(geforce|radeon)\s+`)
	processedStr := re.ReplaceAllString(input, "")
	processedStr = strings.TrimSpace(processedStr)

	// 生成比较用的key
	key := strings.ReplaceAll(strings.ToUpper(processedStr), " ", "")
	if key == "RTX4060TI" {
		var vramSuffix string
		switch vram {
		case 8:
			vramSuffix = "8GB"
		case 16:
			vramSuffix = "16GB"
		}
		fullKey := key + vramSuffix
		if matched, exists := chipsetMap[fullKey]; exists {
			return matched
		}
	}
	if matched, exists := chipsetMap[key]; exists {
		return matched
	} else {
		fmt.Println("unrecognized chipset:", input)
	}
	return input
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
