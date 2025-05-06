package pcData

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type RamSpec struct {
	Code         string
	Brand        string
	Name         string
	Series       string
	Model        string
	Capacity     int
	Type         string
	Speed        int
	Timing       string
	Latency      int
	Voltage      string
	Channel      int
	Profile      string
	LED          string
	HeatSpreader bool
	Prices       []PriceType
	Img          string
}

type RamType struct {
	Id           string
	Brand        string
	Name         string
	NameCN       string
	Series       string
	Model        string
	Capacity     int
	Type         string
	Speed        int
	Timing       string
	Latency      int
	Voltage      string
	Channel      int
	Profile      string
	LED          string
	HeatSpreader bool
	Prices       []PriceType
	Img          string
}

func GetRamSpec(record LinkRecord) RamSpec {
	ramData := RamSpec{}

	if strings.Contains(record.LinkCN, "zol") {
		record.LinkCN = getDetailsLinkFromZol(record.LinkCN, CreateCollector())
	}
	if record.LinkSpec != "" {
		ramData = getRamSpecData(record.LinkSpec, CreateCollector())
	} else if record.LinkUS != "" {
		// newegg
		ramData = getRamUSPrice(record.LinkUS, CreateCollector())
	}

	ramData.Code = record.Name
	ramData.Brand = record.Brand

	if ramData.Name == "" {
		ramData.Name = record.Name
	}
	ramData.Name = RemoveBrandsFromName(ramData.Brand, ramData.Name)

	// 添加各區域價格連結
	ramData.Prices = handleSpecPricesLogic(ramData.Prices, record)
	return ramData
}

func GetRamData(spec RamSpec) (RamType, bool) {
	isValid := true
	newSpec := spec
	nameCN := spec.Name
	collector := CreateCollector()

	// 遍历所有价格数据进行处理
	for _, price := range newSpec.Prices {
		switch price.Region {
		case "CN":
			if strings.Contains(price.PriceLink, "zol") {
				tempSpec := getRamSpecDataFromZol(price.PriceLink, collector)
				newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(RamSpec)

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
				tempSpec := getRamUSPrice(price.PriceLink, collector)
				// 合并图片数据
				if newSpec.Img == "" && tempSpec.Img != "" {
					newSpec.Img = tempSpec.Img
				}
				// 合并规格数据
				newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(RamSpec)

				// 更新价格
				if updatedPrice := getPriceByPlatform(tempSpec.Prices, "US", Platform_Newegg); updatedPrice != nil {
					isValid = isValid && checkPriceValid(updatedPrice.Price)
				}
			}
		}
	}

	channelNum := newSpec.Channel
	if channelNum == 0 {
		channelNum = getRAMChannel(spec.Name)
	}

	// 合并解码结果
	if decodedRAM, ok := DecodeRAMFromPN(spec.Brand, spec.Model); ok {
		decodedRAM.Profile = RamProfileLogic(newSpec, decodedRAM.Profile)
		newSpec = MergeStruct(decodedRAM, newSpec, newSpec.Name).(RamSpec)
	}

	return RamType{
		Id:           SetProductId(spec.Brand, spec.Code),
		Brand:        spec.Brand,
		Name:         spec.Name,
		NameCN:       nameCN,
		Series:       newSpec.Series,
		Model:        newSpec.Model,
		Capacity:     newSpec.Capacity,
		Type:         newSpec.Type,
		Speed:        newSpec.Speed,
		Timing:       newSpec.Timing,
		Latency:      newSpec.Latency,
		Voltage:      newSpec.Voltage,
		Channel:      channelNum,
		LED:          newSpec.LED,
		HeatSpreader: newSpec.HeatSpreader,
		Profile:      newSpec.Profile,
		Prices:       deduplicatePrices(newSpec.Prices),
		Img:          newSpec.Img,
	}, isValid
}

func getRamSpecData(link string, collector *colly.Collector) RamSpec {
	specData := RamSpec{}
	specData.HeatSpreader = false

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner .tns-item img", "src")
		specData.Prices = GetPriceLinkFromPangoly(element)

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Model":
				specData.Model = item.ChildTexts("td")[1]
			case "Speed":
				tempStr := strings.ReplaceAll(item.ChildTexts("td")[1], "-", " ")
				strList := strings.Split(tempStr, " ")
				if strings.Contains(strings.ToUpper(tempStr), "DDR5") {
					specData.Type = "DDR5"
				} else {
					specData.Type = "DDR4"
				}
				if len(strList) > 1 {
					specData.Speed = extractNumberFromString(strList[1])
				}
			case "CAS Latency":
				specData.Latency = extractNumberFromString(item.ChildTexts("td")[1])
			case "Timing":
				specData.Timing = item.ChildTexts("td")[1]
			case "Size":
				sizeList := strings.Split(item.ChildTexts("td")[1], " ")
				specData.Capacity = extractNumberFromString(sizeList[0])
				specData.Channel = getRAMChannel(item.ChildTexts("td")[1])
			case "Voltage":
				specData.Voltage = normalizeVoltage(item.ChildTexts("td")[1])
			case "LED Color":
				specData.LED = item.ChildTexts("td")[1]
			case "Heat Spreader":
				if strings.ToUpper(item.ChildTexts("td")[1]) == "YES" {
					specData.HeatSpreader = true
				}
			}
		})
	})
	collector.Visit(link)

	return specData
}

func getRamUSPrice(link string, collector *colly.Collector) RamSpec {
	specData := RamSpec{}
	specData.HeatSpreader = false

	collectorErrorHandle(collector, link)
	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		specData.Prices = upsertPrice(specData.Prices, PriceType{
			Region:    "US",
			Platform:  Platform_Newegg,
			Price:     extractFloatStringFromString(element.ChildText(".row-side .product-buy-box .price-current")),
			PriceLink: link,
		})

		element.ForEach(".tab-box .tab-panes tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("th") {
			case "Brand":
				specData.Brand = item.ChildText("td")
			case "Series":
				specData.Series = item.ChildText("td")
			case "Model":
				specData.Model = item.ChildText("td")
			case "Capacity":
				specData.Capacity = extractNumberFromString(item.ChildText("td"))
				specData.Channel = getRAMChannel(item.ChildText("td"))
			case "Speed":
				tempStr := strings.ReplaceAll(item.ChildText("td"), "-", " ")
				strList := strings.Split(tempStr, " ")
				if strings.Contains(strings.ToUpper(tempStr), "DDR5") {
					specData.Type = "DDR5"
				} else {
					specData.Type = "DDR4"
				}
				if len(strList) > 1 {
					specData.Speed = extractNumberFromString(strList[1])
				}
			case "CAS Latency":
				specData.Latency = extractNumberFromString(item.ChildText("td"))
			case "Timing":
				specData.Timing = item.ChildText("td")
			case "Voltage":
				specData.Voltage = normalizeVoltage(item.ChildText("td"))
			case "BIOS/Performance Profile":
				specData.Profile = item.ChildText("td")
			case "Heat Spreader":
				if strings.ToUpper(item.ChildText("td")) == "YES" {
					specData.HeatSpreader = true
				}
			}
		})
	})

	collector.Visit(link)
	return specData
}

func getRamSpecDataFromZol(link string, collector *colly.Collector) RamSpec {
	specData := RamSpec{
		HeatSpreader: false,
	}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".wrapper", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".side .goods-card .goods-card__pic img", "src")
		specData.Prices = upsertPrice(specData.Prices, extractJDPriceFromZol(element))
		fmt.Println("specData.Prices : ", specData.Prices)

		element.ForEach(".content table tr", func(i int, item *colly.HTMLElement) {
			convertedHeader := convertGBKString(item.ChildText("th"))
			convertedData := convertGBKString(item.ChildText("td span"))
			// fmt.Println(convertedHeader)
			// fmt.Println(convertedData)

			switch convertedHeader {
			case "发光方式":
				if !strings.Contains(convertedData, "无光") {
					specData.LED = convertedData
				}
			case "内存类型":
				specData.Type = convertedData
			case "容量描述":
				ramNum := 1
				capacity := 0
				totalSize := 0
				if strings.Contains(convertedData, "×") {
					strList := strings.Split(convertedData, "×")
					ramNum = extractNumberFromString(strList[0])
					capacity = extractNumberFromString(strList[1])
					totalSize = ramNum * capacity

					specData.Channel = ramNum
				} else {
					totalSize = extractNumberFromString(convertedData)
					capacity = totalSize
				}
				specData.Capacity = totalSize
			case "CL延迟":
				specData.Latency = extractNumberFromString(convertedData)
				if strings.Contains(convertedData, "-") {
					specData.Timing = convertedData
				}
			case "XMP":
				specData.Profile = "XMP"
			}

			if strings.Contains(convertedHeader, "内存主") {
				specData.Speed = extractNumberFromString(convertedData)
			}
			if strings.Contains(convertedHeader, "散热") {
				specData.HeatSpreader = true
			}
		})

	})
	collector.Visit(link)
	return specData
}

func getRAMChannel(text string) int {
	// 處理括號內 x 數字格式
	if strings.Contains(text, "(") {
		parts := strings.SplitN(text, "(", 2) // 只分割一次
		testStr := parts[1]

		if strings.Contains(testStr, "x") {
			xParts := strings.SplitN(testStr, "x", 2)
			if ramNum := extractNumberFromString(xParts[0]); ramNum > 0 {
				return ramNum
			}
		}
	}

	// 處理關鍵字判斷
	switch {
	case strings.Contains(strings.ToLower(text), "dual"):
		return 2
	case strings.Contains(strings.ToLower(text), "quad"):
		return 4
	default:
		return 1
	}
}

func RamProfileLogic(ram RamSpec, subProfileString string) string {
	// profileList := []string{"Intel XMP 2.0", "Intel XMP 3.0", "AMD EXPO"}
	amdList := []string{"FURY Beast", "Lancer", "Z5 Neo", "银爵", "刃"}
	intelList := []string{"银爵", "刃"}
	isXmp := false
	isExpo := false
	res := ""

	if strContains(ram.Profile, "XMP") || strContains(subProfileString, "XMP") || strings.Contains(strings.ToUpper(ram.Name), " XMP") {
		isXmp = true
	}
	if strContains(ram.Profile, "EXPO") || strContains(subProfileString, "EXPO") || strings.Contains(strings.ToUpper(ram.Name), " AMD") {
		isExpo = true
	}

	if !isXmp {
		for _, item := range intelList {
			if strContains(ram.Series, item) {
				isXmp = true
			}
		}
	}

	if !isExpo {
		for _, item := range amdList {
			if strContains(ram.Series, item) {
				isExpo = true
			}
		}
	}

	if isXmp {
		if ram.Type == "DDR5" {
			res = "Intel XMP 3.0"
		} else {
			res = "Intel XMP 2.0"
		}
	}
	if isExpo {
		if res == "" {
			res = "AMD EXPO"
		} else {
			res += " / AMD EXPO"
		}
	}
	return res
}

func normalizeVoltage(voltageStr string) string {
	// 移除所有空格並轉大寫
	s := strings.ToUpper(strings.ReplaceAll(voltageStr, " ", ""))

	// 直接切割數字部分
	numStr := ""
	for _, c := range s {
		if c == '.' || (c >= '0' && c <= '9') {
			numStr += string(c)
		} else if c == 'V' {
			break // 遇到 V 就停止
		}
	}

	// 轉換為浮點數
	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return ""
	}

	// 智能格式化輸出
	result := strconv.FormatFloat(value, 'f', -1, 64)
	if strings.Contains(result, ".") {
		result = strings.TrimRight(result, "0") // 移除尾部多餘的零
		result = strings.TrimRight(result, ".") // 移除最後剩下的小數點
	}
	return result + "V"
}

var ramSeries = []struct {
	Name     string
	Keywords []string
}{
	// Corsair
	{"Vengeance", []string{"Vengeance"}},
	{"Dominator", []string{"Dominator"}},
	{"Corsair WS", []string{"WS"}}, // 新增工作站專用系列
	{"Corsair LPX", []string{"LPX"}},

	// G.SKILL
	{"Trident Z", []string{"Trident Z", "TridentZ", "Trident Z5", "TridentZ5"}},
	{"Ripjaws", []string{"Ripjaws", "RS5", "Rs5k"}},
	{"Trident Z Royal", []string{"Trident Z Royal"}},
	{"Aegis", []string{"Aegis"}},
	{"Zeta R5", []string{"Zeta R5"}},

	// Kingston
	{"Fury", []string{"Fury"}},
	{"Fury Renegade", []string{"Renegade"}}, // 新增次系列
	{"ValueRAM", []string{"ValueRAM"}},      // 新增經濟型系列
	{"Server Premier", []string{"Server Premier"}},

	// Silicon Power
	{"Xpower", []string{"Xpower"}},
	{"Zenith", []string{"Zenith"}}, // 新增旗艦系列

	// KINGBANK (需處理中文字符)
	{"黑刃", []string{"黑刃", "HeiRen"}},
	{"银爵", []string{"银爵", "YinJue"}},
	{"星刃", []string{"星刃", "XingRen"}},
	{"刃RGB", []string{"刃RGB", "Ren RGB"}},

	// 通用系列
	{"SO-DIMM", []string{"SO-DIMM", "Sodimm"}},
	{"ECC RDIMM", []string{"ECC RDIMM"}},
}

type keywordMapping struct {
	lowerKeyword string
	series       string
}

func DetermineRAMSeries(ramName string) string {
	var keywordMappings []keywordMapping
	for _, series := range ramSeries {
		for _, kw := range series.Keywords {
			keywordMappings = append(keywordMappings, keywordMapping{
				lowerKeyword: strings.ToLower(kw),
				series:       series.Name,
			})
		}
	}

	lowerName := strings.ToLower(ramName)
	for _, mapping := range keywordMappings {
		if strings.Contains(lowerName, mapping.lowerKeyword) {
			return mapping.series
		}
	}
	return "Unknown"
}

// Decode RAM Part Number
func DecodeRAMFromPN(brand, model string) (RamSpec, bool) {
	var ram RamSpec
	isXmp := false
	isExpo := false
	mainPN := strings.ToUpper(strings.ReplaceAll(model, " ", ""))
	lowerBrand := strings.ToLower(brand)

	switch lowerBrand {
	case "corsair":
		mainPN = strings.ToUpper(mainPN)

		// 基本结构校验
		if len(mainPN) < 12 || !strings.HasPrefix(mainPN, "CM") {
			return ram, false
		}

		// 解析系列代码
		seriesCode := mainPN[2:3] // 第3个字符
		seriesMap := map[string]string{
			"U": "Vengeance LED",
			"W": "Vengeance RGB Pro",
			"N": "Vengeance RGB RT",
			"G": "Vengeance RGB RS",
			"D": "Dominator Platinum",
			"K": "Vengeance LPX", // DDR5 Start
			"T": "Dominator Platinum RGB",
			"H": "Vengeance RGB",
			"P": "Dominator Titanium RGB",
		}

		// DDR版本判断
		isXmp = true
		length := len(mainPN)
		lastStr := mainPN[length-5:]
		if strings.Contains(lastStr, "Z") {
			fmt.Println("Corsair AMD EXPO Support :", lastStr)
			isExpo = true
		}

		// 系列名称处理
		if name, exists := seriesMap[seriesCode]; exists {
			ram.Series = name
		}

		// RGB判断逻辑
		switch {
		case strings.Contains(ram.Series, "RGB"):
			ram.LED = "RGB"
		case seriesCode == "U": // Vengeance RGB/Pro系列
			ram.LED = "RGB"
		case seriesCode == "D":
			ram.LED = "White"
		default:
			ram.LED = ""
		}
		// 散热器判断（全系列带散热片）
		ram.HeatSpreader = true
	case "g.skill":
		// 型号示例：F5-6400J3239G16GX2-TZ5RS / F4-3200C16D-16GTZR
		parts := strings.Split(mainPN, "-")
		if len(parts) < 3 {
			return ram, false
		}

		// DDR版本判断
		switch {
		case strings.HasPrefix(parts[0], "F5"):
			ram.Type = "DDR5"
			isXmp = true
			isExpo = true
		case strings.HasPrefix(parts[0], "F4"):
			ram.Type = "DDR4"
			isXmp = true
			isExpo = false
		default:
			return ram, false
		}

		// 解析系列标识
		seriesPart := parts[2]
		var seriesCode string
		// 提取系列代码（取前4个字符，不足则全取）
		if len(seriesPart) >= 4 {
			seriesCode = seriesPart[:4]
		} else {
			seriesCode = seriesPart
		}

		// RGB判断逻辑
		switch {
		case strings.Contains(seriesCode, "TR5N") || // Trident Z5 Royal Neo
			strings.Contains(seriesCode, "TZ5R") || // Trident Z5 RGB
			strings.Contains(seriesCode, "TZ5NR") || // Trident Z5 Neo RGB
			strings.Contains(seriesCode, "RM5R") || // Ripjaws M5 RGB
			strings.Contains(seriesCode, "RM5NR") || // Ripjaws M5 Neo RGB
			strings.Contains(seriesCode, "TZ5CR"): // Trident Z5 CK RGB
			ram.LED = "RGB"
		case strings.HasSuffix(seriesCode, "R") &&
			(strings.HasPrefix(seriesCode, "TR5") ||
				strings.HasPrefix(seriesCode, "TZ5") ||
				strings.HasPrefix(seriesCode, "RM5")):
			ram.LED = "RGB"
		default:
			ram.LED = ""
		}

		// 散热片逻辑（全系列带散热片）
		ram.HeatSpreader = true

		// 系列名称映射（可选，用于更精确的Series字段）
		seriesMap := map[string]string{
			"TR5":   "Trident Z5 Royal",
			"TR5N":  "Trident Z5 Royal Neo",
			"TZ5":   "Trident Z5",
			"TZ5R":  "Trident Z5 RGB",
			"TZ5N":  "Trident Z5 Neo",
			"TZ5NR": "Trident Z5 Neo RGB",
			"RM5R":  "Ripjaws M5 RGB",
			"RM5NR": "Ripjaws M5 Neo RGB",
			"RS5":   "Ripjaws S5",
			"FX5":   "Flare X5",
			"I5":    "Aegis 5",
			"TZ5C":  "Trident Z5 CK",
			"TZ5CR": "Trident Z5 CK RGB",
			"ZR5":   "Zeta R5",
			"ZR5N":  "Zeta R5 Neo",
		}

		// 设置Series字段
		if name, exists := seriesMap[seriesCode]; exists {
			ram.Series = name
		} else {
			// 通用系列判断逻辑
			switch {
			case strings.HasPrefix(seriesCode, "TR5"):
				ram.Series = "Trident Z5 Royal"
			case strings.HasPrefix(seriesCode, "TZ5"):
				ram.Series = "Trident Z5"
			case strings.HasPrefix(seriesCode, "RM5"):
				ram.Series = "Ripjaws M5"
			case strings.HasPrefix(seriesCode, "RS5"):
				ram.Series = "Ripjaws S5"
			case strings.HasPrefix(seriesCode, "FX5"):
				ram.Series = "Flare X5"
			default:
				ram.Series = "Unknown Series"
			}
		}
	case "kingston":
		// 型号示例：KF556C38BBEAK2-32 / KF426C16BB/8
		switch {
		case strings.HasPrefix(mainPN, "KF5"):
			ram.Type = "DDR5"
			isXmp = true
			if len(mainPN) > 10 && mainPN[10] == 'E' {
				isExpo = true
			}
		case strings.HasPrefix(mainPN, "KF4"):
			ram.Type = "DDR4"
			isXmp = true
		}

		if len(mainPN) > 9 && (mainPN[9] == 'B' || mainPN[9] == 'S' || mainPN[9] == 'W') {
			ram.HeatSpreader = true
		}
		if ram.Type == "DDR5" && len(mainPN) > 12 && mainPN[12] == 'A' {
			ram.LED = "RGB"
		}

	case "crucial":
		// 型号示例：BLM2K8G36C16U4B / BLH8G56C46U4B
		switch {
		case strings.Contains(mainPN, "DDR5") || strings.Contains(mainPN, "BLH"):
			ram.Type = "DDR5"
			isXmp = true
			isExpo = true
		case strings.Contains(mainPN, "DDR4") || strings.HasPrefix(mainPN, "BL"):
			ram.Type = "DDR4"
			isXmp = true
		}

		ram.HeatSpreader = strings.HasPrefix(mainPN, "BL")
		if strings.HasSuffix(mainPN, "TF") || strings.Contains(mainPN, "RGB") {
			ram.LED = "RGB"
		}

	case "patriot":
		// 型号示例：PVV432G360C6K / PVR416G360C6K
		switch {
		case strings.HasPrefix(mainPN, "PVV"):
			ram.Type = "DDR5"
			isXmp = true
			isExpo = true
		case strings.HasPrefix(mainPN, "PVN"), strings.HasPrefix(mainPN, "PVR"):
			ram.Type = "DDR4"
			isXmp = true
		}

		ram.HeatSpreader = true
		if strings.Contains(mainPN, "PVR") || strings.Contains(mainPN, "RGB") {
			ram.LED = "RGB"
		}

	case "team group":
		// 型号示例：FF3D532G5600HC36BDC01 / TED48G3200C22DC-S01
		switch {
		case strings.Contains(mainPN, "DDR5") || strings.Contains(mainPN, "D5"):
			isXmp = true
			isExpo = true
		case strings.Contains(mainPN, "DDR4") || strings.Contains(mainPN, "D4"):
			isXmp = true
		}

		ram.HeatSpreader = true
		if strings.Contains(mainPN, "RGB") || strings.Contains(mainPN, "TZR") {
			ram.LED = "RGB"
		}
	default:
		return ram, false
	}

	// 构建Profile信息
	var profiles []string
	if isXmp {
		version := "Intel XMP 3.0"
		if ram.Type == "DDR4" {
			version = "Intel XMP 2.0"
		}
		profiles = append(profiles, version)
	}
	if isExpo {
		profiles = append(profiles, "AMD EXPO")
	}
	ram.Profile = strings.Join(profiles, " / ")

	return ram, true
}
