package pcData

import (
	"fmt"
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
	PriceUS      string
	PriceHK      string
	PriceCN      string
	LinkUS       string
	LinkHK       string
	LinkCN       string
	Img          string
}

type RamType struct {
	Id           string
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
	PriceUS      string
	PriceHK      string
	PriceCN      string
	LinkUS       string
	LinkHK       string
	LinkCN       string
	Img          string
}

func GetRamSpec(record LinkRecord) RamSpec {

	ramData := RamSpec{}
	if strings.Contains(record.LinkCN, "zol") {
		ramData.LinkCN = getDetailsLinkFromZol(record.LinkCN, CreateCollector())
	} else {
		if record.LinkSpec != "" {
			ramData = getRamSpecData(record.LinkSpec, CreateCollector())
		} else if record.LinkUS != "" {
			ramData = getRamUSPrice(record.LinkUS, CreateCollector())
		}
		ramData.PriceCN = record.PriceCN
	}
	ramData.Code = record.Name
	ramData.Brand = record.Brand
	ramData.PriceHK = ""
	ramData.LinkHK = ""
	if record.LinkUS != "" {
		ramData.LinkUS = record.LinkUS
	}
	if ramData.Name == "" {
		ramData.Name = record.Name
	}
	ramData.Name = RemoveBrandsFromName(ramData.Brand, ramData.Name)
	return ramData
}

func GetRamData(spec RamSpec) (RamType, bool) {

	isValid := true

	newSpec := spec

	if strings.Contains(spec.LinkCN, "zol") {
		tempSpec := getRamSpecDataFromZol(spec.LinkCN, CreateCollector())
		// codeStringList := strings.Split(spec.Code, " ")

		newSpec.Img = tempSpec.Img
		if tempSpec.PriceCN != "" {
			newSpec.PriceCN = tempSpec.PriceCN
		}

		newSpec.Series = handleRamSeries(newSpec)
		newSpec.Type = tempSpec.Type
		newSpec.Voltage = tempSpec.Voltage
		newSpec.Capacity = tempSpec.Capacity
		newSpec.Channel = tempSpec.Channel
		newSpec.Timing = tempSpec.Timing
		newSpec.Speed = tempSpec.Speed
		newSpec.Latency = tempSpec.Latency
		newSpec.Profile = tempSpec.Profile
		newSpec.LED = tempSpec.LED
		newSpec.HeatSpreader = tempSpec.HeatSpreader

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

	if strings.Contains(spec.LinkUS, "newegg") {
		newSpec = getRamUSPrice(spec.LinkUS, CreateCollector())

		if newSpec.PriceUS == "" {
			isValid = false
		}
	}

	if spec.PriceCN != "" {
		newSpec.PriceCN = spec.PriceCN
	}

	return RamType{
		Id:           SetProductId(spec.Brand, spec.Code),
		Brand:        spec.Brand,
		Name:         spec.Name,
		Series:       newSpec.Series,
		Model:        newSpec.Model,
		Capacity:     newSpec.Capacity,
		Type:         newSpec.Type,
		Speed:        newSpec.Speed,
		Timing:       newSpec.Timing,
		Latency:      newSpec.Latency,
		Voltage:      newSpec.Voltage,
		Channel:      newSpec.Channel,
		LED:          newSpec.LED,
		HeatSpreader: newSpec.HeatSpreader,
		Profile:      RamProfileLogic(newSpec),
		PriceUS:      newSpec.PriceUS,
		PriceHK:      spec.PriceHK,
		PriceCN:      newSpec.PriceCN,
		LinkHK:       spec.LinkHK,
		LinkUS:       spec.LinkUS,
		LinkCN:       spec.LinkCN,
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
		specData.PriceUS, specData.LinkUS = GetPriceLinkFromPangoly(element)

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
			case "Voltage":
				specData.Voltage = item.ChildTexts("td")[1]
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
		specData.PriceUS = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box .price-current"))

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
				ramNum := 1
				if strings.Contains(item.ChildText("td"), "(") {
					sizeList := strings.Split(item.ChildText("td"), "(")
					specData.Capacity = extractNumberFromString(sizeList[0])
					testStr := sizeList[1]

					if strings.Contains(testStr, "x") {
						secList := strings.Split(testStr, "x")
						ramNum = extractNumberFromString(secList[0])
						specData.Channel = ramNum
					}
				}
				if strings.Contains(strings.ToLower(item.ChildText("td")), "dual") {
					specData.Channel = 2
				} else if strings.Contains(strings.ToLower(item.ChildText("td")), "quad") {
					specData.Channel = 4
				} else {
					specData.Channel = 1
				}

			case "Speed":
				tempStr := strings.ReplaceAll(item.ChildText("td"), "-", " ")
				fmt.Println(tempStr)
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
				specData.Voltage = item.ChildText("td")
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
				fmt.Println(specData.Speed)
			}
			if strings.Contains(convertedHeader, "散热") {
				specData.HeatSpreader = true
			}
		})

	})
	collector.Visit(link)
	return specData
}

func handleRamSeries(spec RamSpec) string {
	if spec.Series == "" {
		if strings.EqualFold(spec.Brand, "kingbank") {
			nameList := strings.Split(spec.Name, " ")
			if len(nameList) > 0 {
				fmt.Println(nameList[0])
				return nameList[0]
			}
		}
	}
	return spec.Series
}

func RamProfileLogic(ram RamSpec) string {
	// profileList := []string{"Intel XMP 2.0", "Intel XMP 3.0", "AMD EXPO"}
	amdList := []string{"FURY Beast", "Lancer", "Z5 Neo", "银爵", "刃"}
	intelList := []string{"银爵", "刃"}
	isXmp := false
	isExpo := false
	res := ""

	fmt.Println(ram.Series)
	fmt.Println(ram.Profile)
	if strContains(ram.Profile, "XMP") {
		isXmp = true
	}
	if strContains(ram.Profile, "EXPO") {
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
			res += ", AMD EXPO"
		}
	}
	return res
}
