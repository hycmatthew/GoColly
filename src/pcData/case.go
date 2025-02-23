package pcData

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
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
	PriceUS            string
	PriceHK            string
	PriceCN            string
	LinkUS             string
	LinkHK             string
	LinkCN             string
	Img                string
}

type CaseType struct {
	Id                 string
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
	PriceUS            string
	PriceHK            string
	PriceCN            string
	LinkUS             string
	LinkHK             string
	LinkCN             string
	Img                string
}

func GetCaseSpec(record LinkRecord) CaseSpec {

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
			"pangoly.com",
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	caseData := CaseSpec{}

	if strings.Contains(record.LinkCN, "zol") {
		caseData.LinkCN = getDetailsLinkFromZol(record.LinkCN, collector)
	} else {
		caseData.LinkCN = record.LinkCN
	}
	if record.LinkSpec != "" {
		caseData = getCaseSpecData(record.LinkSpec, collector)
	}

	caseData.Brand = record.Brand
	caseData.Code = record.Name
	caseData.PriceCN = record.PriceCN
	caseData.PriceHK = ""
	caseData.LinkHK = ""
	if record.LinkUS != "" {
		caseData.LinkUS = record.LinkUS
	}
	if caseData.Name == "" {
		caseData.Name = record.Name
	}
	caseData.Name = RemoveBrandsFromName(caseData.Brand, caseData.Name)
	return caseData
}

func GetCaseData(spec CaseSpec) (CaseType, bool) {

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

	if strings.Contains(spec.LinkCN, "zol") {
		tempSpec := getCaseSpecDataFromZol(spec.LinkCN, cnCollector)

		newSpec := MergeStruct(newSpec, tempSpec).(CaseSpec)
		isValid = checkPriceValid(newSpec.PriceCN)
	}

	if newSpec.PriceCN == "" && strings.Contains(spec.LinkCN, "pconline") {
		newSpec.PriceCN = getCNPriceFromPcOnline(spec.LinkCN, cnCollector)

		isValid = checkPriceValid(newSpec.PriceCN)
	}

	if strings.Contains(spec.LinkUS, "newegg") {
		tempSpec := getCaseUSPrice(spec.LinkUS, usCollector)
		if newSpec.Img == "" {
			tempSpec.Img = newSpec.Img
		}

		newSpec := MergeStruct(tempSpec, newSpec).(CaseSpec)
		isValid = checkPriceValid(newSpec.PriceCN)
	}
	if !isValid {
		fmt.Println(newSpec)
	}

	return CaseType{
		Id:                 SetProductId(spec.Brand, spec.Code),
		Brand:              spec.Brand,
		Name:               spec.Name,
		ReleaseDate:        spec.ReleaseDate,
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
		PriceUS:            newSpec.PriceUS,
		PriceHK:            "",
		PriceCN:            newSpec.PriceCN,
		LinkUS:             spec.LinkUS,
		LinkHK:             spec.LinkHK,
		LinkCN:             spec.LinkCN,
		Img:                newSpec.Img,
	}, isValid
}

func getCaseSpecData(link string, collector *colly.Collector) CaseSpec {
	specData := CaseSpec{}
	collectorErrorHandle(collector, link)

	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner .tns-item img", "src")
		specData.PriceUS, specData.LinkUS = GetPriceLinkFromPangoly(element)

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
		specData.PriceUS = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box .price-current"))
		available := element.ChildText(".row-side .product-buy-box .product-buy .btn-message")
		specData.PriceUS = OutOfStockLogic(specData.PriceUS, available)

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

func CompareCaseDataLogic(cur CaseType, list []CaseType) CaseType {
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
