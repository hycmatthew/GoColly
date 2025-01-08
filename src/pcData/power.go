package pcData

import (
	"fmt"
	"net/http"
	"net/url"
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

	if record.LinkSpec == "" {
		powerData.LinkCN = getDetailsLinkFromZol(record.LinkCN, specCollector)
	} else {
		powerData := getPowerSpecData(record.LinkSpec, specCollector)
		powerData.LinkCN = record.LinkCN
	}

	powerData.Code = record.Name
	powerData.Name = record.Name
	powerData.Brand = record.Brand
	powerData.PriceCN = record.PriceCN
	powerData.PriceHK = ""
	powerData.LinkHK = ""
	if record.LinkUS != "" {
		powerData.LinkUS = record.LinkUS
	}
	return powerData
}

func GetPowerData(spec PowerSpec) (PowerType, bool) {

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

	priceCN := spec.PriceCN
	if spec.LinkCN != "" {
		tempSpec := getPowerSpecDataFromZol(spec.LinkCN, cnCollector)

		spec.Img = tempSpec.Img
		spec.Standard = tempSpec.Standard
		spec.Wattage = tempSpec.Wattage
		spec.Size = tempSpec.Size
		spec.Modular = tempSpec.Modular
		spec.Efficiency = tempSpec.Efficiency
		spec.Length = tempSpec.Length
		priceCN = tempSpec.PriceCN

		if priceCN == "" {
			isValid = false
		}
	}
	priceUS, tempImg := spec.PriceUS, spec.Img
	if strings.Contains(spec.LinkUS, "newegg") {
		priceUS, tempImg = getUSPriceAndImgFromNewEgg(spec.LinkUS, usCollector)

		if priceUS == "" {
			isValid = false
		}
	}

	return PowerType{
		Brand:       spec.Brand,
		Name:        spec.Name,
		ReleaseDate: spec.ReleaseDate,
		Standard:    spec.Standard,
		Wattage:     spec.Wattage,
		Size:        spec.Size,
		Modular:     spec.Modular,
		Efficiency:  spec.Efficiency,
		Length:      spec.Length,
		LinkUS:      spec.LinkUS,
		LinkHK:      spec.LinkHK,
		LinkCN:      spec.LinkCN,
		PriceCN:     priceCN,
		PriceUS:     priceUS,
		PriceHK:     "",
		Img:         tempImg,
	}, isValid
}

func getPowerSpecData(link string, collector *colly.Collector) PowerSpec {
	specData := PowerSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner img", "src")
		loopBreak := false

		element.ForEach("table.table-prices tr", func(i int, item *colly.HTMLElement) {
			if !loopBreak {
				specData.PriceUS = extractFloatStringFromString(item.ChildText(".detail-purchase strong"))
				tempLink := item.ChildAttr(".detail-purchase", "href")

				if strings.Contains(tempLink, "amazon") {
					amazonLink := strings.Split(tempLink, "?tag=")[0]
					specData.LinkUS = amazonLink
					loopBreak = true
				}
				if strings.Contains(tempLink, "newegg") {
					neweggLink := strings.Split(tempLink, "url=")[1]
					UnescapeLink, _ := url.QueryUnescape(neweggLink)
					specData.LinkUS = strings.Split(UnescapeLink, "\u0026")[0]
					loopBreak = true
				}
			}
		})

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
			fmt.Println(convertedHeader)

			switch convertedHeader {
			case "电源版本":
				specData.Standard = convertedData
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
