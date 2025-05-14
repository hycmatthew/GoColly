package pcData

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type SSDSpec struct {
	Code        string
	Brand       string
	Name        string
	ReleaseDate string
	Model       string
	Capacity    string
	MaxRead     int
	MaxWrite    int
	Read4K      int
	Write4K     int
	Interface   string
	FlashType   string
	FormFactor  string
	Prices      []PriceType
	Img         string
}

type SSDType struct {
	Id          string
	Brand       string
	Name        string
	NameCN      string
	ReleaseDate string
	Model       string
	Capacity    string
	MaxRead     int
	MaxWrite    int
	Read4K      int
	Write4K     int
	Interface   string
	FlashType   string
	FormFactor  string
	Prices      []PriceType
	Img         string
}

var ssdScoreList []SSDSpec

func init() {
	ssdScoreList = getDetailsLinkFrommSSDTester(CreateCollector())
	fmt.Println("SSD数据初始化完成")
	fmt.Println("SSD数据数量: ", len(ssdScoreList))
}

func cleanText(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\u00a0", " ")) // 替换&nbsp;
}

func extractNumber(s string) int {
	re := regexp.MustCompile(`[\d,]+`)
	numStr := re.FindString(s)
	numStr = strings.ReplaceAll(numStr, ",", "")
	num, _ := strconv.Atoi(numStr)
	return num
}

func getDetailsLinkFrommSSDTester(collector *colly.Collector) []SSDSpec {
	ssds := []SSDSpec{}
	link := "https://ssd-tester.com/top_ssd.php"

	collectorErrorHandle(collector, link)
	collector.OnHTML("#table", func(element *colly.HTMLElement) {
		element.ForEach("tbody tr", func(i int, item *colly.HTMLElement) {
			var ssd SSDSpec
			tds := item.DOM.ChildrenFiltered("td")

			// 解析品牌和名称
			brandName := strings.TrimSpace(tds.Eq(1).Text())
			if split := strings.SplitN(brandName, " ", 2); len(split) > 1 {
				ssd.Brand = split[0]
				ssd.Name = split[1]
			}

			// 解析图片
			ssd.Img, _ = tds.Eq(2).Find("img").Attr("src")

			// 容量
			ssd.Capacity = cleanText(tds.Eq(3).Text())

			// 闪存类型
			ssd.FlashType = cleanText(tds.Eq(4).Text())

			// 接口类型
			ssd.Interface = cleanText(tds.Eq(7).Text())

			// 性能数据
			ssd.MaxRead = extractNumber(tds.Eq(9).Text())
			ssd.MaxWrite = extractNumber(tds.Eq(10).Text())
			ssd.Read4K = extractNumber(tds.Eq(11).Text()) // 根据实际HTML结构调整索引
			ssd.Write4K = extractNumber(tds.Eq(12).Text())

			ssds = append(ssds, ssd)
		})
	})
	collector.Visit(link)
	return ssds
}

func GetSSDSpec(record LinkRecord) SSDSpec {
	ssdData := SSDSpec{}

	// 处理中国区链接
	if strings.Contains(record.LinkCN, "zol") {
		record.LinkCN = getDetailsLinkFromZol(record.LinkCN, CreateCollector())
	}

	if record.LinkSpec != "" {
		ssdData = getSSDSpecData(record.LinkSpec, CreateCollector())
	}

	ssdData.Code = record.Name
	ssdData.Brand = record.Brand
	if ssdData.Name == "" {
		ssdData.Name = record.Name
	}

	// 尝试从预加载的评分数据匹配
	tempName := ssdData.Brand + " " + RemoveBrandsFromName(ssdData.Brand, ssdData.Name)
	baseName, targetCapacityGB := parseNameAndCapacity(tempName)
	closestSpec, found := findClosestSpec(baseName, targetCapacityGB)

	if found {
		fmt.Println("找到匹配的SSD规格: ", closestSpec.Name, " - ", record.Name)
		ssdData.FlashType = closestSpec.FlashType
		ssdData.Interface = closestSpec.Interface
		ssdData.Read4K = closestSpec.Read4K
		ssdData.Write4K = closestSpec.Write4K
	}
	ssdData.Name = RemoveBrandsFromName(ssdData.Brand, ssdData.Name)

	// 添加各區域價格連結
	ssdData.Prices = handleSpecPricesLogic(ssdData.Prices, record)
	return ssdData
}

func GetSSDData(spec SSDSpec) (SSDType, bool) {
	isValid := true
	newSpec := spec
	nameCN := spec.Name
	collector := CreateCollector()

	// 遍历所有价格数据进行处理
	for _, price := range newSpec.Prices {
		switch price.Region {
		case "CN":
			if strings.Contains(price.PriceLink, "zol") {
				tempSpec := getSSDSpecDataFromZol(price.PriceLink, collector)
				newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(SSDSpec)

				// 更新价格信息
				if updatedPrice := getPriceByPlatform(tempSpec.Prices, "CN", Platform_JD); updatedPrice != nil {
					isValid = isValid && checkPriceValid(updatedPrice.Price)
				}
			}
			if strings.Contains(price.PriceLink, "pconline") {
				if price.Price == "" {
					tempNameCN, priceCN := getCNNameAndPriceFromPcOnline(price.PriceLink, collector)
					nameCN = tempNameCN
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
				priceUS, tempImg := getUSPriceAndImgFromNewEgg(price.PriceLink, collector)
				if tempImg != "" {
					newSpec.Img = tempImg
				}
				fmt.Println(spec.Name, " orginPrice: ", price, "- priceUS: ", priceUS)
				newSpec.Prices = upsertPrice(newSpec.Prices, PriceType{
					Region:    "US",
					Platform:  Platform_Newegg,
					Price:     priceUS,
					PriceLink: price.PriceLink,
				})
				isValid = isValid && checkPriceValid(priceUS)
			}
		}
	}

	return SSDType{
		Id:          SetProductId(spec.Brand, spec.Code),
		Brand:       newSpec.Brand,
		Name:        newSpec.Name,
		NameCN:      nameCN,
		ReleaseDate: newSpec.ReleaseDate,
		Model:       newSpec.Model,
		Capacity:    NormalizeSSDCapacity(newSpec.Capacity),
		MaxRead:     newSpec.MaxRead,
		MaxWrite:    newSpec.MaxWrite,
		Read4K:      newSpec.Read4K,
		Write4K:     newSpec.Write4K,
		Interface:   NormalizeSSDInterface(newSpec.Interface),
		FlashType:   newSpec.FlashType,
		FormFactor:  newSpec.FormFactor,
		Prices:      deduplicatePrices(newSpec.Prices),
		Img:         newSpec.Img,
	}, isValid
}

func getSSDSpecData(link string, collector *colly.Collector) SSDSpec {
	specData := SSDSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {

		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner img", "src")
		specData.Prices = GetPriceLinkFromPangoly(element)

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Model":
				specData.Model = strings.Split(item.ChildTexts("td")[1], "\n")[0]
			case "Release Date":
				specData.ReleaseDate = item.ChildText("td span")
			case "Capacity":
				specData.Capacity = item.ChildTexts("td")[1]
			case "Interface":
				specData.Interface = item.ChildTexts("td")[1]
			case "Form Factor":
				specData.FormFactor = item.ChildTexts("td")[1]
			case "NAND Flash Type":
				specData.FlashType = item.ChildTexts("td")[1]
			case "Max Sequential Read":
				specData.MaxRead = extractNumberFromString(item.ChildTexts("td")[1])
			case "Max Sequential Write":
				specData.MaxWrite = extractNumberFromString(item.ChildTexts("td")[1])
			case "4KB Random Read":
				specData.Read4K = extractNumberFromString(item.ChildTexts("td")[1])
			case "4KB Random Write":
				specData.Write4K = extractNumberFromString(item.ChildTexts("td")[1])
			}
		})
	})

	collector.Visit(link)

	return specData
}

func getSSDSpecDataFromZol(link string, collector *colly.Collector) SSDSpec {
	specData := SSDSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".wrapper", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".side .goods-card .goods-card__pic img", "src")
		fmt.Println(element.DOM.Html())
		specData.Prices = upsertPrice(specData.Prices, extractJDPriceFromZol(element))

		element.ForEach(".content table tr", func(i int, item *colly.HTMLElement) {
			convertedHeader := convertGBKString(item.ChildText("th"))
			convertedData := convertGBKString(item.ChildText("td span"))

			switch convertedHeader {
			case "存储容量":
				specData.Capacity = convertedData
			case "接口类型":
				specData.FormFactor = convertedData
			case "读取速度":
				specData.MaxRead = extractNumberFromString(convertedData)
			case "写入速度":
				specData.MaxWrite = extractNumberFromString(convertedData)
			case "4K随机":
				if specData.Read4K == 0 {
					specData.Read4K = extractNumberFromString(convertedData)
				} else {
					specData.Write4K = extractNumberFromString(convertedData)
				}
			case "通道":
				if strings.Contains(convertedData, "x4") {
					specData.Interface = "PCIe " + convertedData
				} else {
					specData.Interface = convertedData
				}
			}

			if strings.Contains(convertedHeader, "架构") {
				specData.FlashType = convertedData
			}

		})

	})
	collector.Visit(link)
	return specData
}

func getZhiTaiDataFromPcOnline(link string, collector *colly.Collector) SSDSpec {
	specData := SSDSpec{
		FormFactor: "M.2-2280",
		Interface:  "PCle Gen 3x4",
	}

	collectorErrorHandle(collector, link)

	collector.OnHTML(".product-detail-main", func(element *colly.HTMLElement) {
		mallPrice := extractFloatStringFromString(element.ChildText(".product-price-info .product-mallSales em.price"))
		// otherPrice := extractFloatStringFromString(element.ChildText(".product-price-info .product-price-other span"))
		normalPrice := extractFloatStringFromString(element.ChildText(".product-price-info .r-price a"))

		specData.Prices = upsertPrice(specData.Prices, PriceType{
			Region:    "CN",
			Platform:  Platform_JD,
			Price:     firstNonEmpty(mallPrice, normalPrice, ""),
			PriceLink: "",
		})

		element.ForEach(".baseParam dd i", func(i int, item *colly.HTMLElement) {
			convertedString := convertGBKString(item.Text)
			if strings.Contains(convertedString, "类型") {
				dataStrList := strings.Split(string(convertedString), "：")
				specData.FlashType = dataStrList[len(dataStrList)-1]
			}
			if strings.Contains(convertedString, "连续读取") {
				specData.MaxRead = extractNumberFromString(convertedString)
			}
			if strings.Contains(convertedString, "连续写入") {
				specData.MaxWrite = extractNumberFromString(convertedString)
			}
		})

		if specData.MaxRead > 4000 {
			specData.Interface = "PCle Gen4X4"
		}
	})

	collector.Visit(link)
	return specData
}

func NormalizeSSDCapacity(input string) string {
	// 正则表达式匹配容量数值和单位
	re := regexp.MustCompile(`(?i)^\s*([\d.]+)\s*([TGMK]?B?)\s*$`)
	matches := re.FindStringSubmatch(input)
	if matches == nil {
		fmt.Println("无效格式: ", input)
		return input
	}

	valueStr := matches[1]
	unit := strings.ToUpper(strings.TrimSpace(matches[2]))

	// 转换为浮点数
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		fmt.Println("数值解析失败: ", input)
		return input
	}

	// 转换为统一GB单位（按1TB=1000GB换算）
	var totalGB float64
	switch {
	case strings.HasPrefix(unit, "TB") || unit == "T":
		totalGB = value * 1000
	case strings.HasPrefix(unit, "GB") || unit == "G":
		totalGB = value
	case strings.HasPrefix(unit, "MB") || unit == "M":
		totalGB = value / 1000
	case strings.HasPrefix(unit, "KB") || unit == "K":
		totalGB = value / 1e6
	default:
		fmt.Println("未知单位: ", input)
		return ""
	}

	// 特殊值处理映射表（GB基准值 => 标准格式）
	capacityMap := map[int]string{
		500:  "500GB",
		1000: "1TB",
		2000: "2TB",
		3000: "3TB", // 根据需求可调整
		4000: "4TB",
		8000: "8TB",
	}

	// 特殊阈值处理（单位：GB）
	switch {
	case totalGB >= 490 && totalGB < 750: // 500GB ±10%
		return capacityMap[500]
	case totalGB >= 980 && totalGB < 1500: // 1TB ±2%
		return capacityMap[1000]
	case totalGB >= 1960 && totalGB < 2500: // 2TB ±10%
		return capacityMap[2000]
	case totalGB >= 3840 && totalGB < 4200: // 4TB ±5%
		return capacityMap[4000]
	case totalGB >= 7680 && totalGB < 8500: // 8TB ±5%
		return capacityMap[8000]
	}

	// 精确匹配标准容量
	if formatted, exists := capacityMap[int(totalGB)]; exists {
		return formatted
	}
	return input
}

func NormalizeSSDInterface(input string) string {
	// 统一预处理
	cleaned := strings.ToUpper(input)
	cleaned = regexp.MustCompile(`[\s\-]+`).ReplaceAllString(cleaned, "") // 移除空格和连字符

	// 主匹配正则表达式
	re := regexp.MustCompile(`(?i)(PCI[E]?|SATA|NVME?)(?:([GEN]*)(\d+\.?\d*)|(\d+\.?\d*)[A-Z]*)(?:X(\d+)|(\d+)X|X(\d+))?`)

	matches := re.FindStringSubmatch(cleaned)
	if matches == nil {
		fmt.Println("无效格式: ", input)
		return ""
	}

	// 解析组件
	var protocol, version, lanes string

	// 协议类型处理
	switch {
	case strings.Contains(matches[1], "PCI"):
		protocol = "PCIe"
	case strings.Contains(matches[1], "SATA"):
		protocol = "SATA"
	case strings.Contains(matches[1], "NVME"):
		protocol = "NVMe" // NVMe单独处理
	}

	// 版本号处理 (支持 Gen4/4.0/4 等格式)
	version = extractVersion(matches[2:5])

	// 通道数处理 (x4/4x/×4)
	lanes = extractLanes(matches[5:8])

	// 特殊处理NVMe
	if protocol == "NVMe" {
		return formatNVMe(version, lanes)
	}

	// 组合标准格式
	return formatStandard(protocol, version, lanes)
}

// 辅助函数：提取版本号
func extractVersion(matchGroups []string) string {
	for _, g := range matchGroups {
		if g != "" {
			ver := regexp.MustCompile(`(\d+\.?\d*)`).FindString(g)
			if ver != "" {
				if !strings.Contains(ver, ".") {
					return ver + ".0"
				}
				return ver
			}
		}
	}
	return "3.0" // 默认版本
}

// 辅助函数：提取通道数
func extractLanes(matchGroups []string) string {
	for _, g := range matchGroups {
		if g != "" {
			return "x" + g
		}
	}
	return "" // 无通道数时不显示
}

// 格式化NVMe接口
func formatNVMe(version, lanes string) string {
	base := "NVMe PCIe"
	if version != "" {
		base += " " + version
	}
	if lanes != "" {
		base += lanes
	}
	return base
}

// 格式化标准接口
func formatStandard(protocol, version, lanes string) string {
	parts := []string{protocol}

	if protocol == "SATA" {
		if version != "" {
			parts = append(parts, version)
		}
	} else {
		if version != "" {
			parts = append(parts, version)
		}
		if lanes != "" {
			parts = append(parts, lanes)
		}
	}

	return strings.Join(parts, " ")
}

// 核心匹配逻辑
// 新增辅助函数：从产品名称解析基础名称和容量
func parseNameAndCapacity(name string) (baseName string, capacityGB int) {
	re := regexp.MustCompile(`(?i)\s*(\d+)\s*(TB|GB)\s*$`)
	matches := re.FindStringSubmatch(name)
	if len(matches) == 3 {
		capacity, _ := strconv.Atoi(matches[1])
		unit := strings.ToUpper(matches[2])
		if unit == "TB" {
			capacityGB = capacity * 1000
		} else {
			capacityGB = capacity
		}
		baseName = strings.TrimSuffix(name, matches[0])
		baseName = strings.ToUpper(strings.TrimSpace(baseName))
		return
	}
	return name, 0
}

// 新增辅助函数：解析容量字符串为GB数值
func parseCapacityToGB(capacityStr string) int {
	cleaned := cleanText(capacityStr)
	re := regexp.MustCompile(`(?i)(\d+)\s*(TB|GB)`)
	matches := re.FindStringSubmatch(cleaned)
	if len(matches) < 3 {
		return 0
	}
	value, _ := strconv.Atoi(matches[1])
	unit := strings.ToUpper(matches[2])
	if unit == "TB" {
		return value * 1000
	}
	return value
}

// 新增辅助函数：寻找最接近的规格
func findClosestSpec(baseName string, targetCapacityGB int) (SSDSpec, bool) {
	fmt.Println("寻找最接近的SSD规格: ", baseName, " - ", targetCapacityGB)
	var candidates []SSDSpec
	for _, spec := range ssdScoreList {
		// fmt.Println("当前规格: ", spec.Name, " - ", spec.Capacity)
		specBase, _ := parseNameAndCapacity(spec.Name)
		if specBase == baseName {
			candidates = append(candidates, spec)
			continue
		}
		// 额外检查规格名称是否包含基础名称（更宽松的匹配）
		if strings.Contains(specBase, baseName) || strings.Contains(baseName, specBase) {
			candidates = append(candidates, spec)
		}
	}

	if len(candidates) == 0 {
		return SSDSpec{}, false
	}
	fmt.Println("找到候选规格数量: ", len(candidates))
	// 优先寻找完全匹配型号
	for _, spec := range candidates {
		if spec.Name == baseName {
			return spec, true
		}
	}

	// 找最接近容量的规格
	closest := candidates[0]
	closestDiff := abs(targetCapacityGB - parseCapacityToGB(closest.Capacity))
	for _, spec := range candidates[1:] {
		currentCapacity := parseCapacityToGB(spec.Capacity)
		currentDiff := abs(targetCapacityGB - currentCapacity)
		if currentDiff < closestDiff || currentDiff == closestDiff {
			closest = spec
			closestDiff = currentDiff
		}
	}
	return closest, true
}

// 绝对值函数
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
