package pcData

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type PowerSpec struct {
	Code        string
	Brand       string
	Name        string
	ReleaseDate string
	Wattage     int
	Size        string
	Standard    string
	Modular     string
	Efficiency  string
	Length      int
	PriceUS     string
	PriceHK     string
	PriceCN     string
	LinkUS      string
	LinkHK      string
	LinkCN      string
	Img         string
}

type PowerType struct {
	Id          string
	Brand       string
	Name        string
	ReleaseDate string
	Wattage     int
	Size        string
	Standard    string
	Modular     string
	Efficiency  string
	Length      int
	PriceUS     string
	PriceHK     string
	PriceCN     string
	LinkUS      string
	LinkHK      string
	LinkCN      string
	Img         string
}

func GetPowerSpec(record LinkRecord) PowerSpec {

	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"www.newegg.com",
			"newegg.com",
			"www.price.com.hk",
			"price.com.hk",
			"detail.zol.com.cn",
			"zol.com.cn",
			"pangoly.com",
			"www.huntkey.cn",
			"huntkey.cn",
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	specCollector := collector.Clone()
	powerData := PowerSpec{}

	if strings.Contains(record.LinkCN, "zol") {
		powerData.LinkCN = getDetailsLinkFromZol(record.LinkCN, collector)
	} else {
		powerData.LinkCN = record.LinkCN
	}
	if record.LinkSpec != "" {
		powerData = getPowerSpecData(record.LinkSpec, specCollector)
	}

	powerData.Code = record.Name
	powerData.Brand = record.Brand
	powerData.PriceCN = record.PriceCN
	if record.LinkUS != "" {
		powerData.LinkUS = record.LinkUS
	}
	if powerData.Name == "" {
		powerData.Name = record.Name
	}
	powerData.Name = RemoveBrandsFromName(powerData.Brand, powerData.Name)
	return powerData
}

func GetPowerData(spec PowerSpec) (PowerType, bool) {
	fmt.Println(SetProductId(spec.Brand, spec.Code))
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

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})
	cnCollector := collector.Clone()
	usCollector := collector.Clone()
	isValid := true

	newSpec := spec
	if spec.LinkCN != "" {
		tempSpec := getPowerSpecDataFromZol(spec.LinkCN, cnCollector)

		newSpec := MergeStruct(newSpec, tempSpec, newSpec.Name).(PowerSpec)
		isValid = checkPriceValid(newSpec.PriceCN)
	}

	if strings.Contains(spec.LinkUS, "newegg") {
		tempSpec := getPowerUSPrice(spec.LinkUS, usCollector)
		tempSpec.Standard = comparePSUStandard(tempSpec.Standard, newSpec.Standard)
		if newSpec.Img == "" {
			tempSpec.Img = newSpec.Img
		}

		newSpec := MergeStruct(tempSpec, newSpec, newSpec.Name).(PowerSpec)
		isValid = checkPriceValid(newSpec.PriceCN)
	}

	return PowerType{
		Id:          SetProductId(spec.Brand, spec.Code),
		Brand:       newSpec.Brand,
		Name:        newSpec.Name,
		ReleaseDate: newSpec.ReleaseDate,
		Standard:    newSpec.Standard,
		Wattage:     newSpec.Wattage,
		Size:        newSpec.Size,
		Modular:     newSpec.Modular,
		Efficiency:  newSpec.Efficiency,
		Length:      newSpec.Length,
		LinkUS:      newSpec.LinkUS,
		LinkHK:      newSpec.LinkHK,
		LinkCN:      newSpec.LinkCN,
		PriceCN:     newSpec.PriceCN,
		PriceUS:     newSpec.PriceUS,
		PriceHK:     newSpec.PriceHK,
		Img:         newSpec.Img,
	}, isValid
}

func getPowerSpecData(link string, collector *colly.Collector) PowerSpec {
	specData := PowerSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner img", "src")
		specData.PriceUS, specData.LinkUS = GetPriceLinkFromPangoly(element)
		fmt.Println(specData.LinkUS)

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Release Date":
				specData.ReleaseDate = item.ChildText("td span")
			case "Wattage":
				specData.Wattage = extractNumberFromString(item.ChildTexts("td")[1])
			case "Type":
				specData.Size = item.ChildTexts("td")[1]
			case "Modular":
				specData.Modular = item.ChildTexts("td")[1]
			case "Efficiency":
				specData.Efficiency = item.ChildTexts("td")[1]
			case "Length":
				specData.Length = extractNumberFromString(item.ChildTexts("td")[1])
			}
		})
	})

	collector.Visit(link)

	return specData
}

func getPowerUSPrice(link string, collector *colly.Collector) PowerSpec {
	specData := PowerSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		specData.PriceUS = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box .price-current"))
		available := element.ChildText(".row-side .product-buy-box .product-buy .btn-message")
		specData.PriceUS = OutOfStockLogic(specData.PriceUS, available)

		prdName := element.ChildText(".product-title")
		standard := extractATXStandard(prdName)

		element.ForEach(".tab-box .tab-panes tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("th") {
			case "Type":
				standard = comparePSUStandard(standard, extractATXStandard(item.ChildText("td")))
			}
		})
		specData.Standard = standard
	})

	collector.Visit(link)
	return specData
}

func getPowerSpecDataFromZol(link string, collector *colly.Collector) PowerSpec {
	specData := PowerSpec{
		Standard: "ATX 2.0",
	}

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

			switch convertedHeader {
			case "电源版本":
				specData.Standard = extractATXStandard(convertedData)
			case "电源模组":
				if strings.Contains(convertedData, "全模组") {
					specData.Modular = "Full"
				} else if strings.Contains(convertedData, "半模组") {
					specData.Modular = "Semi"
				} else {
					specData.Modular = "No"
				}
			case "额定功率":
				specData.Wattage = extractNumberFromString(convertedData)
			case "80PLUS认证":
				if strings.Contains(convertedData, "钛金") {
					specData.Efficiency = "80+ Titanium"
				} else if strings.Contains(convertedData, "白金") || strings.Contains(convertedData, "铂金") {
					specData.Efficiency = "80+ Platinum"
				} else if strings.Contains(convertedData, "金牌") {
					specData.Efficiency = "80+ Gold"
				} else if strings.Contains(convertedData, "银牌") {
					specData.Efficiency = "80+ Silver"
				} else if strings.Contains(convertedData, "铜") {
					specData.Efficiency = "80+ Bronze"
				} else {
					specData.Efficiency = convertedData
				}
			}

			if strings.Contains(convertedHeader, "效率") {
				if specData.Efficiency == "" {
					specData.Efficiency = convertedData
				}
			}

			if strings.Contains(convertedHeader, "电源尺") {
				sizeStr := strings.Split(convertedData, "×")
				if strings.Contains(sizeStr[0], "125") {
					specData.Size = "SFX"
					sizeStr = removeElement(sizeStr, "125")
				} else {
					specData.Size = "ATX"
					sizeStr = removeElement(sizeStr, "150")
				}
				specData.Length = getTheLargestValueInArray(sizeStr)
			}
		})

	})
	collector.Visit(link)
	return specData
}

/*
func getPowerSpecDataFromHuntkey(link string, collector *colly.Collector) PowerSpec {
	specData := PowerSpec{
		Modular: "Full",
	}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".border_hui", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".lbProName")
		specData.Img = element.ChildAttr("#deta_pic .act img", "src")

		if strings.Contains(specData.Name, "JUMPER") || strings.Contains(specData.Name, "战鹰") || strings.Contains(specData.Name, "狼牙") {
			specData.Modular = "No"
		}
		if strings.HasPrefix(specData.Name, "ECO") || strings.HasPrefix(specData.Name, "GS") || strings.HasPrefix(specData.Name, "WD") || strings.HasPrefix(specData.Name, "GX") || strings.Contains(specData.Name, "猛擎") {
			specData.Modular = "No"
		}
		if strings.Contains(specData.Name, "MVP") {
			if specData.Name == "MVP500" || specData.Name == "MVP600" {
				specData.Modular = "Semi"
			} else {
				specData.Modular = "Full"
			}
		}
		if strings.HasPrefix(specData.Name, "HYPER") || strings.HasPrefix(specData.Name, "IP") || strings.Contains(specData.Name, "重火力") {
			specData.Modular = "Full"
		}
		if strings.Contains(specData.Name, "全模组") {
			specData.Modular = "Full"
		}

		fmt.Println(specData.Name)
		fmt.Println(specData.Modular)

		var mergedStrList []string

		element.ForEach("#con_one_2 tr td", func(i int, item *colly.HTMLElement) {
			if i%2 == 0 {
				mergedStrList = append(mergedStrList, item.Text)
			} else {
				setIndex := (i - 1) / 2
				mergedStrList[setIndex] = mergedStrList[setIndex] + item.Text
			}
		})

		for _, item := range mergedStrList {
			valStr := strings.Split(item, "：")[1]
			if strings.Contains(item, "额定功率") {
				specData.Wattage = extractNumberFromString(valStr)
			}
			if strings.Contains(item, "转换效率") {
				if strings.Contains(item, "80PLUS铜牌") {
					specData.Efficiency = "80+ Bronze"
				} else if strings.Contains(item, "80PLUS银牌") {
					specData.Efficiency = "80+ Silver"
				} else if strings.Contains(item, "80PLUS金牌") {
					specData.Efficiency = "80+ Gold"
				} else if strings.Contains(item, "80PLUS白金") || strings.Contains(item, "80PLUS铂金") {
					specData.Efficiency = "80+ Platinum"
				} else if strings.Contains(item, "80PLUS钛金") {
					specData.Efficiency = "80+ Titanium"
				} else {
					specData.Efficiency = valStr
				}
			}
			if strings.Contains(item, "尺寸") {
				sizeStr := strings.Split(item, "*")
				if strings.Contains(sizeStr[0], "125") {
					specData.Size = "SFX"
				} else {
					specData.Size = "ATX"
				}
				specData.Length = extractNumberFromString(sizeStr[2])
			}
		}
	})

	collector.Visit(link)
	return specData
}

func getPowerSpecDataFromPcOnline(link string, collector *colly.Collector) PowerSpec {
	detailsLink := strings.Split(link, ".htm")[0] + "_detail.html"
	specData := PowerSpec{}

	collectorErrorHandle(collector, detailsLink)
	collector.OnHTML(".area-detailparams", func(element *colly.HTMLElement) {
		element.ForEach(".bd-box tr", func(i int, item *colly.HTMLElement) {
			convertedHeader := convertGBKString(item.ChildText("th"))
			convertedData := convertGBKString(item.ChildText("td"))
			convertedLinkData := convertGBKString(item.ChildText("td .poptxt"))

			switch convertedHeader {
			case "型号":
				specData.Name = convertedData
			case "电源标准":
				specData.Size = convertedData
			case "额定功率":
				specData.Wattage = extractNumberFromString(convertedData)
			case "80PLUS认证":
				if strings.Contains(convertedLinkData, "铜牌") {
					specData.Efficiency = "80+ Bronze"
				} else if strings.Contains(convertedLinkData, "银牌") {
					specData.Efficiency = "80+ Silver"
				} else if strings.Contains(convertedLinkData, "金牌") {
					specData.Efficiency = "80+ Gold"
				} else if strings.Contains(convertedLinkData, "白金") || strings.Contains(convertedLinkData, "铂金") {
					specData.Efficiency = "80+ Platinum"
				} else if strings.Contains(convertedLinkData, "钛金") {
					specData.Efficiency = "80+ Titanium"
				} else {
					specData.Efficiency = convertedLinkData
				}
			}

			if strings.Contains(convertedHeader, "尺") {
				fmt.Println(convertedData)
				sizeStr := strings.Split(convertedData, "×")
				fmt.Println(sizeStr)
				specData.Length = extractNumberFromString(sizeStr[len(sizeStr)-1])
			}
		})

	})
	collector.Visit(detailsLink)
	return specData
}
*/

func ComparePowerDataLogic(cur PowerType, list []PowerType) PowerType {
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

func extractATXStandard(name string) string {
	// 將名稱轉為大寫並去除空格
	normalized := strings.ToUpper(strings.ReplaceAll(name, " ", ""))

	// 定義正則表達式規則
	patterns := map[string]string{
		"ATX 3.1": `ATX(?:12V)?3\.1`,
		"ATX 3.0": `ATX(?:12V)?3\.0`,
		"ATX 3":   `ATX(?:12V)?3$`,
	}

	// 按順序匹配正則表達式
	for standard, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, normalized)
		if matched {
			return standard
		}
	}
	// 如果沒有匹配，返回默認的 ATX 2.0
	return "ATX 2.0"
}

func comparePSUStandard(str1 string, str2 string) string {
	if str1 == "ATX 2.0" {
		return str2
	}
	return str1
}
