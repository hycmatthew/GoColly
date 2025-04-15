package pcData

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

type CaseSpec struct {
	Code               string
	Brand              string
	Name               string
	ReleaseDate        string
	Color              string
	CaseSize           string
	PowerSupply        bool
	DriveBays2         int
	DriveBays3         int
	Compatibility      []string
	Dimensions         []int
	MaxVGAlength       int
	RadiatorSupport    int
	MaxCpuCoolorHeight int
	SlotsNum           int
	Prices             []PriceType
	Img                string
}

type CaseType struct {
	Id                 string
	Brand              string
	Name               string
	NameCN             string
	ReleaseDate        string
	Color              string
	CaseSize           string
	PowerSupply        bool
	DriveBays2         int
	DriveBays3         int
	Compatibility      []string
	Dimensions         []int
	MaxVGAlength       int
	RadiatorSupport    int
	MaxCpuCoolorHeight int
	SlotsNum           int
	Prices             []PriceType
	Img                string
}

func GetCaseSpec(record LinkRecord) CaseSpec {
	caseData := CaseSpec{}
	// 处理zol链接
	if strings.Contains(record.LinkCN, "zol") {
		record.LinkCN = getDetailsLinkFromZol(record.LinkCN, CreateCollector())
	}

	// 获取规格数据
	if record.LinkSpec != "" {
		caseData = getCaseSpecData(record.LinkSpec, CreateCollector())
	}
	caseData.Brand = record.Brand
	caseData.Code = record.Name

	// 处理名称
	if caseData.Name == "" {
		caseData.Name = record.Name
	}
	caseData.Name = RemoveBrandsFromName(caseData.Brand, caseData.Name)

	// 添加各區域價格連結
	caseData.Prices = handleSpecPricesLogic(caseData.Prices, record)
	return caseData
}

func GetCaseData(spec CaseSpec) (CaseType, bool) {
	isValid := true
	newSpec := spec
	nameCN := spec.Name
	collector := CreateCollector()

	// 遍历所有价格数据进行处理
	for _, price := range newSpec.Prices {
		// 根據平台類型進行處理
		switch price.Region {
		case "CN":
			if strings.Contains(price.PriceLink, "zol") {
				tempSpec := getCaseSpecDataFromZol(price.PriceLink, collector)
				newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(CaseSpec)

				// 更新價格信息
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
			tempSpec := getCaseUSPrice(price.PriceLink, collector)
			// 合併圖片數據
			if newSpec.Img == "" && tempSpec.Img != "" {
				newSpec.Img = tempSpec.Img
			}
			// 合併其他規格數據
			newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(CaseSpec)

			// 更新價格
			if updatedPrice := getPriceByPlatform(tempSpec.Prices, "US", Platform_Newegg); updatedPrice != nil {
				isValid = isValid && checkPriceValid(updatedPrice.Price)
			}
		}
	}

	return CaseType{
		Id:                 SetProductId(newSpec.Brand, spec.Code),
		Brand:              newSpec.Brand,
		Name:               newSpec.Name,
		NameCN:             nameCN,
		ReleaseDate:        newSpec.ReleaseDate,
		CaseSize:           newSpec.CaseSize,
		Color:              newSpec.Color,
		PowerSupply:        newSpec.PowerSupply,
		DriveBays2:         newSpec.DriveBays2,
		DriveBays3:         newSpec.DriveBays3,
		Compatibility:      newSpec.Compatibility,
		Dimensions:         newSpec.Dimensions,
		MaxVGAlength:       newSpec.MaxVGAlength,
		RadiatorSupport:    newSpec.RadiatorSupport,
		MaxCpuCoolorHeight: newSpec.MaxCpuCoolorHeight,
		SlotsNum:           newSpec.SlotsNum,
		Prices:             deduplicatePrices(newSpec.Prices),
		Img:                newSpec.Img,
	}, isValid
}

func getCaseSpecData(link string, collector *colly.Collector) CaseSpec {
	specData := CaseSpec{}
	collectorErrorHandle(collector, link)

	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner .tns-item img", "src")
		specData.Prices = GetPriceLinkFromPangoly(element)

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Release Date":
				specData.ReleaseDate = item.ChildText("td span")
			case "Type":
				specData.CaseSize = item.ChildTexts("td")[1]
			case "Color":
				specData.Color = item.ChildTexts("td")[1]
			case "Includes Power Supply":
				if item.ChildTexts("td")[1] != "No" {
					specData.PowerSupply = true
				}
			case `Internal 2.5" Drive Bays`:
				specData.DriveBays2 = extractNumberFromString(item.ChildTexts("td")[1])
			case `Internal 3.5" Drive Bays`:
				specData.DriveBays3 = extractNumberFromString(item.ChildTexts("td")[1])
			case "Motherboard Compatibility":
				sizeList := strings.Split(item.ChildTexts("td")[1], ",")
				for i, item := range sizeList {
					sizeList[i] = GetFormFactorLogic(strings.TrimSpace(item))
				}
				specData.Compatibility = RemoveDuplicates(sizeList)
			case "Dimensions":
				tempDimensions := strings.Split(item.ChildTexts("td")[1], "x")
				var dimensionsList []int
				counter := 0
				for _, item := range tempDimensions {
					if counter < 3 {
						dimensionsList = append(dimensionsList, extractNumberFromString(item))
						counter++
					}
				}
				specData.Dimensions = dimensionsList
			case "Max VGA length allowance":
				specData.MaxVGAlength = extractNumberFromString(item.ChildTexts("td")[1])
			case "Expansion Slots":
				specData.SlotsNum = extractNumberFromString(item.ChildTexts("td")[1])
			}
		})
	})

	collector.Visit(link)
	return specData
}

func getCaseUSPrice(link string, collector *colly.Collector) CaseSpec {
	specData := CaseSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		tempPrice := extractFloatStringFromString(element.ChildText(".row-side .product-buy-box .price-current"))
		available := element.ChildText(".row-side .product-buy-box .product-buy .btn-message")
		tempPrice = OutOfStockLogic(tempPrice, available)

		specData.Prices = upsertPrice(specData.Prices, PriceType{
			Region:    "US",
			Platform:  Platform_Newegg,
			Price:     tempPrice,
			PriceLink: link,
		})

		element.ForEach(".tab-box .tab-panes tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("th") {
			case "Radiator Options":
				tempStrList := strings.Split(item.ChildText("td"), ":")
				highestSize := 120
				fmt.Println(tempStrList)
				for _, item := range tempStrList {
					tempSize := extractNumberFromString(item)

					if tempSize > highestSize {
						fmt.Println(tempSize)
						highestSize = tempSize
					}
				}
				specData.RadiatorSupport = highestSize
			case "Max CPU Cooler Height":
				specData.MaxCpuCoolorHeight = extractNumberFromString(item.ChildText("td"))
			}
		})
	})

	collector.Visit(link)
	return specData
}

func getCaseSpecDataFromZol(link string, collector *colly.Collector) CaseSpec {
	specData := CaseSpec{
		PowerSupply: false,
	}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".wrapper", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".side .goods-card .goods-card__pic img", "src")
		specData.Prices = upsertPrice(specData.Prices, extractJDPriceFromZol(element))

		element.ForEach(".content table tr", func(i int, item *colly.HTMLElement) {
			convertedHeader := convertGBKString(item.ChildText("th"))
			convertedData := convertGBKString(item.ChildText("td span"))
			// fmt.Println(convertedHeader)
			// fmt.Println(convertedData)

			switch convertedHeader {
			case "适用主板":
				compatibilityStr := strings.Split(convertedData, "，")
				var compatibilityArr []string
				for _, item := range compatibilityStr {
					tempCompatibility := GetFormFactorLogic(item)
					compatibilityArr = append(compatibilityArr, tempCompatibility)
				}
				specData.Compatibility = RemoveDuplicates(compatibilityArr)
			case "扩展插槽":
				specData.SlotsNum = extractNumberFromString(convertedData)
			case "颜色":
				specData.Color = convertedData
			case "显卡限长":
				specData.MaxVGAlength = extractNumberFromString(convertedData)
			case "80PLUS认证":
				if strings.Contains(convertedData, "钛金") {
					specData.CaseSize = "80+ Titanium"
				} else if strings.Contains(convertedData, "白金") || strings.Contains(convertedData, "铂金") {
					specData.CaseSize = "80+ Platinum"
				} else if strings.Contains(convertedData, "金牌") {
					specData.CaseSize = "80+ Gold"
				} else if strings.Contains(convertedData, "银牌") {
					specData.CaseSize = "80+ Silver"
				} else if strings.Contains(convertedData, "铜") {
					specData.CaseSize = "80+ Bronze"
				} else {
					specData.CaseSize = convertedData
				}
			}

			if strings.Contains(convertedHeader, "水冷") {
				tempStrList := SplitAny(convertedData, "/×")
				highestSize := 120

				for _, item := range tempStrList {
					fmt.Println(item)
					tempSize := extractNumberFromString(item)
					fmt.Println(tempSize)
					if tempSize > highestSize {
						highestSize = tempSize
					}
				}
				specData.RadiatorSupport = highestSize
				fmt.Println("RadiatorSupport : ", highestSize)
			}
			if strings.Contains(convertedHeader, "结构") {
				specData.CaseSize = convertedData
			}
			if strings.Contains(convertedHeader, "CPU散热器") {
				specData.MaxCpuCoolorHeight = extractNumberFromString(convertedData)
			}
			if strings.Contains(convertedHeader, "2.5英") {
				specData.DriveBays2 = extractNumberFromString(convertedData)
			}
			if strings.Contains(convertedHeader, "3.5英") {
				specData.DriveBays3 = extractNumberFromString(convertedData)
			}

			if strings.Contains(convertedHeader, "产品尺") {
				tempDimensions := SplitAny(convertedData, "x×*")
				var dimensionsList []int
				counter := 0
				for _, item := range tempDimensions {
					if counter < 3 {
						dimensionsList = append(dimensionsList, extractNumberFromString(item))
						counter++
					}
				}

				specData.Dimensions = dimensionsList
			}
		})

	})
	collector.Visit(link)
	return specData
}
