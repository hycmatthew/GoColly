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
	PriceUS     string
	PriceHK     string
	PriceCN     string
	LinkUS      string
	LinkHK      string
	LinkCN      string
	Img         string
}

type SSDType struct {
	Id          string
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
	PriceUS     string
	PriceHK     string
	PriceCN     string
	LinkUS      string
	LinkHK      string
	LinkCN      string
	Img         string
}

func GetSSDSpec(record LinkRecord) SSDSpec {

	ssdData := SSDSpec{}
	if strings.Contains(record.LinkCN, "zol") {
		ssdData.LinkCN = getDetailsLinkFromZol(record.LinkCN, CreateCollector())
	} else {
		if record.LinkSpec != "" {
			ssdData = getSSDSpecData(record.LinkSpec, CreateCollector())
		}
		ssdData.LinkCN = record.LinkCN
	}

	ssdData.Code = record.Name
	ssdData.Brand = record.Brand
	ssdData.PriceCN = record.PriceCN
	ssdData.PriceHK = ""
	ssdData.LinkHK = ""
	if record.LinkUS != "" {
		ssdData.LinkUS = record.LinkUS
	}
	if ssdData.Name == "" {
		ssdData.Name = record.Name
	}
	ssdData.Name = RemoveBrandsFromName(ssdData.Brand, ssdData.Name)
	return ssdData
}

func GetSSDData(spec SSDSpec) (SSDType, bool) {
	isValid := true

	newSpec := spec

	if strings.Contains(spec.LinkCN, "zol") {
		tempSpec := getSSDSpecDataFromZol(spec.LinkCN, CreateCollector())
		// codeStringList := strings.Split(spec.Code, " ")

		newSpec.Img = tempSpec.Img
		if tempSpec.PriceCN != "" {
			newSpec.PriceCN = tempSpec.PriceCN
		}

		newSpec.Capacity = tempSpec.Capacity
		newSpec.FlashType = tempSpec.FlashType
		newSpec.FormFactor = tempSpec.FormFactor
		newSpec.MaxRead = tempSpec.MaxRead
		newSpec.MaxWrite = tempSpec.MaxWrite
		newSpec.Interface = tempSpec.Interface

		if newSpec.PriceCN == "" {
			isValid = false
		}
	}

	if newSpec.PriceCN == "" && strings.Contains(spec.LinkCN, "pconline") {
		newSpec.PriceCN = getCNPriceFromPcOnline(spec.LinkCN, CreateCollector())

		if newSpec.PriceCN == "" {
			isValid = false
		}
	}

	tempImg := ""
	if strings.Contains(spec.LinkUS, "newegg") {
		newSpec.PriceUS, tempImg = getUSPriceAndImgFromNewEgg(spec.LinkUS, CreateCollector())

		if tempImg != "" {
			newSpec.Img = tempImg
		}
		if newSpec.PriceUS == "" {
			isValid = false
		}
	}

	if spec.PriceCN != "" {
		newSpec.PriceCN = spec.PriceCN
	}

	return SSDType{
		Id:          SetProductId(spec.Brand, spec.Code),
		Brand:       spec.Brand,
		Name:        spec.Name,
		ReleaseDate: newSpec.ReleaseDate,
		Model:       newSpec.Model,
		Capacity:    NormalizeSSDCapacity(newSpec.Capacity),
		MaxRead:     newSpec.MaxRead,
		MaxWrite:    newSpec.MaxWrite,
		Interface:   NormalizeSSDInterface(newSpec.Interface),
		FlashType:   newSpec.FlashType,
		FormFactor:  newSpec.FormFactor,
		PriceUS:     newSpec.PriceUS,
		PriceHK:     "",
		PriceCN:     newSpec.PriceCN,
		LinkUS:      spec.LinkUS,
		LinkHK:      spec.LinkHK,
		LinkCN:      spec.LinkCN,
		Img:         newSpec.Img,
	}, isValid
}

func getSSDSpecData(link string, collector *colly.Collector) SSDSpec {
	specData := SSDSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {

		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner img", "src")
		specData.PriceUS, specData.LinkUS = GetPriceLinkFromPangoly(element)

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

		mallPrice := extractFloatStringFromString(element.ChildText("side .goods-card .item-b2cprice span"))
		// otherPrice := extractFloatStringFromString(element.ChildText(".price__merchant .price"))
		normalPrice := extractFloatStringFromString(element.ChildText(".side .goods-card .goods-card__price span"))
		if mallPrice != "" {
			specData.PriceCN = mallPrice
		} else {
			specData.PriceCN = normalPrice
		}

		element.ForEach(".content table tr", func(i int, item *colly.HTMLElement) {
			convertedHeader := convertGBKString(item.ChildText("th"))
			convertedData := convertGBKString(item.ChildText("td span"))
			fmt.Println(convertedHeader)
			fmt.Println(convertedData)

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

		otherPrice := extractFloatStringFromString(element.ChildText(".product-price-info .product-price-other span"))

		normalPrice := extractFloatStringFromString(element.ChildText(".product-price-info .r-price a"))

		if mallPrice != "" {
			specData.PriceCN = mallPrice
		} else if otherPrice != "" {
			specData.PriceCN = otherPrice
		} else {
			specData.PriceCN = normalPrice
		}

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
