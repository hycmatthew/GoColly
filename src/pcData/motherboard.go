package pcData

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type MotherboardSpec struct {
	Code        string
	Name        string
	Brand       string
	Socket      string
	Chipset     string
	RamSlot     int
	RamType     string
	RamSupport  []int
	RamMax      int
	Pcie16Slot  int
	Pcie4Slot   int
	Pcie1Slot   int
	PcieSlotStr []string
	M2Slot      int
	SataSlot    int
	FormFactor  string
	Wireless    bool
	Prices      []PriceType // 替换原有PriceXX和LinkXX
	Img         string
}

type MotherboardType struct {
	Id          string
	Name        string
	NameCN      string
	Brand       string
	Socket      string
	Chipset     string
	RamSlot     int
	RamType     string
	RamSupport  []int
	RamMax      int
	Pcie16Slot  int
	Pcie4Slot   int
	Pcie1Slot   int
	PcieSlotStr []string
	M2Slot      int
	SataSlot    int
	FormFactor  string
	Wireless    bool
	Prices      []PriceType // 替换原有PriceXX和LinkXX
	Img         string
}

func GetMotherboardSpec(record LinkRecord) MotherboardSpec {
	motherboardData := MotherboardSpec{}
	fmt.Println(record.LinkSpec)
	if strings.Contains(record.LinkSpec, "asus.com") {
		motherboardData = getMotherboardSpecDataFromAsus(record.LinkSpec, CreateCollector())
	} else if strings.Contains(record.LinkSpec, "msi.com") {
		motherboardData = getMotherboardSpecDataFromMsi(record.LinkSpec, CreateCollector())
	} else if strings.Contains(record.LinkSpec, "pangoly.com") {
		motherboardData = getMotherboardSpecData(record.LinkSpec, CreateCollector())
	}

	if strings.Contains(strings.ToUpper(record.Name), "WIFI") {
		motherboardData.Wireless = true
	}

	motherboardData.Code = record.Name
	motherboardData.Brand = record.Brand
	motherboardData.Name = RemoveBrandsFromName(motherboardData.Brand, motherboardData.Name)

	// 添加各區域價格連結
	motherboardData.Prices = handleSpecPricesLogic(motherboardData.Prices, record)
	return motherboardData
}

func GetMotherboardData(spec MotherboardSpec) (MotherboardType, bool) {
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
				newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(MotherboardSpec)

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
				if tempImg == "404" {
					continue
				}
				if tempImg != "" {
					newSpec.Img = tempImg
				}
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

	return MotherboardType{
		Id:          SetProductId(newSpec.Brand, newSpec.Code),
		Name:        newSpec.Name,
		NameCN:      nameCN,
		Brand:       newSpec.Brand,
		Socket:      newSpec.Socket,
		Chipset:     normalizeMotherboardChipset(newSpec.Chipset),
		RamSlot:     newSpec.RamSlot,
		RamType:     newSpec.RamType,
		RamSupport:  newSpec.RamSupport,
		RamMax:      newSpec.RamMax,
		Pcie16Slot:  newSpec.Pcie16Slot,
		Pcie4Slot:   newSpec.Pcie4Slot,
		Pcie1Slot:   newSpec.Pcie1Slot,
		PcieSlotStr: newSpec.PcieSlotStr,
		M2Slot:      newSpec.M2Slot,
		SataSlot:    newSpec.SataSlot,
		FormFactor:  GetFormFactorLogic(newSpec.FormFactor),
		Wireless:    newSpec.Wireless,
		Prices:      deduplicatePrices(newSpec.Prices),
		Img:         newSpec.Img,
	}, isValid
}

func getMotherboardSpecData(link string, collector *colly.Collector) MotherboardSpec {
	specData := MotherboardSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".tns-inner img", "src")
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Prices = GetPriceLinkFromPangoly(element)

		specData.RamType = element.ChildText(".table-striped .badge-primary")
		var ramSupportList []int

		element.ForEach(".table-striped .ram-values span", func(i int, item *colly.HTMLElement) {
			temp := extractNumberFromString(strings.Replace(item.Text, "", "", -1))
			ramSupportList = append(ramSupportList, temp)
		})

		specData.RamSupport = ramSupportList
		var slotList []string

		element.ForEach("ul.tail-links a", func(i int, item *colly.HTMLElement) {
			itemStr := item.ChildText("strong")
			if strings.Contains(itemStr, "PCI-Express x16 Slots") {
				specData.Pcie16Slot = extractNumberFromString(itemStr)
				tempSlotStr := strconv.Itoa(specData.Pcie16Slot) + " PCI-Express x16 Slot"
				if specData.Pcie16Slot > 1 {
					tempSlotStr = strconv.Itoa(specData.Pcie16Slot) + " PCI-Express x16 Slots"
				}
				slotList = append(slotList, tempSlotStr)
			}
			if strings.Contains(itemStr, "PCI-Express x4 Slots") {
				specData.Pcie4Slot = extractNumberFromString(itemStr)
				tempSlotStr := strconv.Itoa(specData.Pcie16Slot) + " PCI-Express x4 Slot"
				if specData.Pcie16Slot > 1 {
					tempSlotStr = strconv.Itoa(specData.Pcie16Slot) + " PCI-Express x4 Slots"
				}
				slotList = append(slotList, tempSlotStr)
			}
			if strings.Contains(itemStr, "PCI-Express x1 Slots") {
				specData.Pcie1Slot = extractNumberFromString(itemStr)
				tempSlotStr := strconv.Itoa(specData.Pcie16Slot) + " PCI-Express x1 Slot"
				if specData.Pcie16Slot > 1 {
					tempSlotStr = strconv.Itoa(specData.Pcie16Slot) + " PCI-Express x1 Slots"
				}
				slotList = append(slotList, tempSlotStr)
			}
			if strings.Contains(itemStr, "M.2 Ports") {
				specData.M2Slot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "RAM Slots") {
				specData.RamSlot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "Supported RAM") {
				specData.RamMax = extractNumberFromString(itemStr)
			}
		})
		specData.PcieSlotStr = slotList

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Socket":
				specData.Socket = item.ChildTexts("td")[1]
			case "Form factor":
				specData.FormFactor = item.ChildTexts("td")[1]
			case "Chipset":
				specData.Chipset = item.ChildTexts("td")[1]
			}
		})
	})

	collector.Visit(link)

	return specData
}

func getMotherboardSpecDataFromAsus(link string, collector *colly.Collector) MotherboardSpec {
	specData := MotherboardSpec{}

	collectorErrorHandle(collector, link)

	collector.OnHTML(".TechSpec__section__9V8DZ", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".TechSpec__rowImage__35vd6 img", "src")
		specData.Name = element.ChildText(".TechSpec__itemName__an9aU .pdName")

		element.ForEach(".TechSpec__rowTable__1LR9D", func(i int, item *colly.HTMLElement) {
			itemStr := item.ChildText(".rowTableTitle")
			dataStr := item.ChildText(".rowTableItemViewBox div")

			if strings.Contains(itemStr, "CPU") {
				tempStrList := strings.Split(dataStr, " ")
				for _, item := range tempStrList {
					if strings.Contains(item, "LGA") {
						specData.Socket = item
						break
					}
					if strings.Contains(item, "AM4") || strings.Contains(item, "AM5") {
						specData.Socket = item
						break
					}
				}
			}
			if strings.Contains(itemStr, "Chipset") {
				specData.Chipset = dataStr
			}
			if strings.Contains(itemStr, "Memory") {
				var ramSupportList []int
				ramStrList := strings.Split(dataStr, " ")
				for i, item := range ramStrList {
					if i == 0 {
						specData.RamSlot = extractNumberFromString(item)
					}
					if strings.Contains(item, "GB") {
						specData.RamMax = extractNumberFromString(item)
					}
					if strings.Contains(item, "DDR4") || strings.Contains(item, "DDR5") {
						specData.RamType = item
					}
					ramTestList := getAllRamSupportList()

					for _, speedItem := range ramTestList {
						if strings.Contains(item, speedItem) {
							ramSupportList = append(ramSupportList, extractNumberFromString(speedItem))
						}
					}
				}
				specData.RamSupport = ramSupportList
			}
			if strings.Contains(itemStr, "Expansion Slots") {
				replacements := map[string]string{
					"<strong>":   "",
					"</strong>":  "",
					"<sup>":      "",
					"</sup>":     "",
					"<u>":        "",
					"</u>":       "",
					"<nil>":      "",
					"\u0026amp;": "&",
				}
				item.ForEach(".TechSpec__rowTableItems__KYWXp", func(i int, expansionItem *colly.HTMLElement) {
					expansionItemStr, _ := expansionItem.DOM.Html()
					for oldStr, newStr := range replacements {
						expansionItemStr = strings.Replace(expansionItemStr, oldStr, newStr, -1)
					}
					expansionStrList := strings.Split(expansionItemStr, "<br/>")
					var expansionResList []string

					for _, item := range expansionStrList {
						if strings.Contains(item, " slot") {
							expansionResList = append(expansionResList, item)
							if strings.Contains(item, "x16 slot") {
								specData.Pcie16Slot += 1
							}
							if strings.Contains(item, "x4 slot") {
								specData.Pcie4Slot += 1
							}
							if strings.Contains(item, "x1 slot") {
								specData.Pcie1Slot += 1
							}
						}
					}
					specData.PcieSlotStr = expansionResList
				})
			}
			if strings.Contains(itemStr, "Storage") {
				dataStr := item.ChildText(".rowTableItemViewBox div strong")
				specData.M2Slot = extractNumberFromString(getWordBeforeSpecificString(dataStr, "x M.2 slots"))
				specData.SataSlot = extractNumberFromString(getWordBeforeSpecificString(dataStr, "x SATA"))
			}
			if strings.Contains(itemStr, "Form Factor") {
				dataStr := item.ChildText(".rowTableItemViewBox div")
				formFactorStrList := strings.Split(dataStr, " ")
				specData.FormFactor = formFactorStrList[0]
			}
		})
	})

	collector.Visit(link)

	return specData
}

func getMotherboardSpecDataFromMsi(link string, collector *colly.Collector) MotherboardSpec {
	specData := MotherboardSpec{}

	collectorErrorHandle(collector, link)

	collector.OnHTML(".product-mainbox", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".specContainer .img-container img", "src")
		specData.Name = element.ChildText(".text-center h3")

		element.ForEach(".spec-block-div .table tr", func(i int, item *colly.HTMLElement) {
			itemStr := item.ChildText("th span")
			dataStr := item.ChildText("td")
			addSocketLogic := false

			if strings.Contains(itemStr, "CPU") {
				tempStrList := strings.Split(dataStr, " ")
				for _, item := range tempStrList {
					if addSocketLogic {
						specData.Socket = "LGA" + item
					}
					if strings.Contains(item, "LGA") {
						if item == "LGA" {
							addSocketLogic = true
						} else {
							specData.Socket = item
						}
						break
					}
					if strings.Contains(item, "AM4") || strings.Contains(item, "AM5") {
						specData.Socket = item
						break
					}
				}
			}
			if strings.Contains(itemStr, "Chipset") {
				specData.Chipset = dataStr
			}
			if strings.Contains(itemStr, "Memory") {
				var ramSupportList []int
				ramStrList := strings.Split(dataStr, " ")
				for i, item := range ramStrList {
					if i == 0 {
						specData.RamSlot = extractNumberFromString(item)
					}
					if strings.Contains(item, "GB") {
						specData.RamMax = extractNumberFromString(item)
					}
					if strings.Contains(item, "DDR4") || strings.Contains(item, "DDR5") {
						specData.RamType = strings.ReplaceAll(item, ",", "")
					}
					ramTestList := getAllRamSupportList()

					for _, speedItem := range ramTestList {
						if strings.Contains(item, speedItem) {
							ramSupportList = append(ramSupportList, extractNumberFromString(speedItem))
						}
					}
				}
				specData.RamSupport = ramSupportList
			}
			if strings.Contains(itemStr, "Slot") {
				item.ForEach("td", func(i int, expansionItem *colly.HTMLElement) {
					expansionItemStr, _ := expansionItem.DOM.Html()
					expansionStrList := strings.Split(expansionItemStr, "<br/>")
					var expansionResList []string

					for _, item := range expansionStrList {
						if strings.Contains(item, " slot") {
							expansionResList = append(expansionResList, strings.TrimSpace(item))
							if strings.Contains(item, "x16 slot") {
								specData.Pcie16Slot += 1
							}
							if strings.Contains(item, "x4 slot") {
								specData.Pcie4Slot += 1
							}
							if strings.Contains(item, "x1 slot") {
								specData.Pcie1Slot += 1
							}
						}
					}
					specData.PcieSlotStr = expansionResList
				})
			}
			if strings.Contains(itemStr, "Storage") {
				specData.M2Slot = extractNumberFromString(getWordBeforeSpecificString(dataStr, "x M.2"))
				specData.SataSlot = extractNumberFromString(getWordBeforeSpecificString(dataStr, "x SATA"))
			}
			if strings.Contains(itemStr, "PCB Info") {
				specData.FormFactor = dataStr
			}
		})
	})

	collector.Visit(link)

	return specData
}

func getAllRamSupportList() []string {
	return []string{
		"2133", "2400", "2666", "2800", "2933",
		"3000", "3200", "3333", "3400", "3466", "3600", "3733", "3866",
		"4000", "4133", "4266", "4400", "4500", "4600", "4700", "4800",
		"5000", "5066", "5133", "5200", "5333", "5400", "5600", "5800",
		"6000", "6200", "6400", "6600", "6800",
		"7000", "7200", "7400", "7600", "7800",
		"8000", "8200", "8400", "8600", "8800",
		"9000", "9200",
	}
}

func GetFormFactorLogic(str string) string {
	formFactorList := []string{"mATX", "M-ATX", "Micro-ATX", "Micro ATX", "Mini-ITX", "Mini ITX", "ITX", "EATX", "E-ATX"}
	upperStr := strings.ToUpper(str)
	for _, item := range formFactorList {
		if strings.Contains(upperStr, strings.ToUpper(item)) {
			switch item {
			case "mATX", "M-ATX", "Micro-ATX", "Micro ATX":
				return "Micro-ATX"
			case "Mini-ITX", "Mini ITX":
				return "Mini-ITX"
			case "ITX":
				return "ITX"
			case "EATX", "E-ATX":
				return "EATX"
			}
		}
	}
	return "ATX"
}

func normalizeMotherboardChipset(chipset string) string {
	// 預定義有效芯片組列表
	validChipsets := map[string]bool{
		"Z890":  true,
		"B860":  true,
		"H810":  true,
		"W790":  true,
		"Z790":  true,
		"B760":  true,
		"Z690":  true,
		"Q670":  true,
		"H670":  true,
		"B660":  true,
		"H610":  true,
		"C621A": false,
		// AMD
		"X870E": true,
		"X870":  true,
		"B850":  true,
		"B840":  true,
		"X670E": true,
		"X670":  true,
		"B650E": true,
		"B650":  true,
		"A620":  true,
		"X570":  true,
		"B550":  true,
		"A520":  true,
	}

	// 清理並標準化輸入字符串
	cleaned := strings.ToUpper(chipset)

	// 使用正則表達式提取芯片組型號
	re := regexp.MustCompile(`([A-Z])(\d{3}[A-Z]?)`)
	matches := re.FindStringSubmatch(cleaned)

	var extracted string
	if len(matches) >= 2 {
		// 組合字母和數字部分
		extracted = matches[1] + matches[2]
	} else {
		// 備用匹配模式（例如處理類似 "H610" 的格式）
		reSimple := regexp.MustCompile(`([A-Z]\d{3}[A-Z]?)`)
		extracted = reSimple.FindString(cleaned)
	}

	// 有效性檢查
	if extracted == "" {
		fmt.Printf("無法識別芯片組格式: %s\n", chipset)
		return ""
	}

	if !validChipsets[extracted] {
		fmt.Printf("發現未定義芯片組: %s (原始輸入: %s)\n", extracted, chipset)
		return "" // 或者返回 extracted 根據需求
	}

	return extracted
}
