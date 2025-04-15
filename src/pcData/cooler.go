package pcData

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

type CoolerSpec struct {
	Code             string
	Brand            string
	Name             string
	ReleaseDate      string
	Sockets          []string
	IsLiquidCooler   bool
	LiquidCoolerSize int
	AirCoolerHeight  int
	NoiseLevel       string
	FanSpeed         string
	Airflow          string
	Pressure         string
	LED              string
	Prices           []PriceType
	Img              string
}

type CoolerType struct {
	Id               string
	Brand            string
	Name             string
	NameCN           string
	ReleaseDate      string
	Sockets          []string
	IsLiquidCooler   bool
	LiquidCoolerSize int
	AirCoolerHeight  int
	NoiseLevel       string
	FanSpeed         string
	Airflow          string
	Pressure         string
	LED              string
	Prices           []PriceType
	Img              string
}

func GetCoolerSpec(record LinkRecord) CoolerSpec {
	coolerData := CoolerSpec{}

	if strings.Contains(record.LinkCN, "zol") {
		record.LinkCN = getDetailsLinkFromZol(record.LinkCN, CreateCollector())
	}

	if record.LinkSpec != "" {
		if strings.Contains(record.LinkSpec, "zol") {
			coolerData = getCoolerSpecDataFromZol(record.LinkSpec, CreateCollector())
		} else {
			coolerData = getCoolerSpecData(record.LinkSpec, CreateCollector())
		}
	}

	coolerData.Code = record.Name
	coolerData.Brand = record.Brand
	if coolerData.Name == "" {
		coolerData.Name = record.Name
	}
	coolerData.Name = RemoveBrandsFromName(coolerData.Brand, coolerData.Name)

	// 合并价格链接
	coolerData.Prices = handleSpecPricesLogic(coolerData.Prices, record)
	return coolerData
}

func GetCoolerData(spec CoolerSpec) (CoolerType, bool) {
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
				tempSpec := getCoolerSpecDataFromZol(price.PriceLink, collector)
				newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(CoolerSpec)

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
			priceUS, tempImg := getUSPriceAndImgFromNewEgg(price.PriceLink, collector)
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

	// 自动检测LED类型
	if strings.Contains(newSpec.Name, "RGB") {
		newSpec.LED = "RGB"
	} else if strings.Contains(newSpec.Name, "ARGB") {
		newSpec.LED = "ARGB"
	}

	return CoolerType{
		Id:               SetProductId(newSpec.Brand, spec.Code),
		Brand:            newSpec.Brand,
		Name:             newSpec.Name,
		NameCN:           nameCN,
		ReleaseDate:      newSpec.ReleaseDate,
		Sockets:          newSpec.Sockets,
		IsLiquidCooler:   newSpec.IsLiquidCooler,
		AirCoolerHeight:  newSpec.AirCoolerHeight,
		LiquidCoolerSize: newSpec.LiquidCoolerSize,
		NoiseLevel:       newSpec.NoiseLevel,
		FanSpeed:         newSpec.FanSpeed,
		Airflow:          newSpec.Airflow,
		Pressure:         newSpec.Pressure,
		LED:              newSpec.LED,
		Prices:           deduplicatePrices(newSpec.Prices),
		Img:              newSpec.Img,
	}, isValid
}

func getCoolerSpecData(link string, collector *colly.Collector) CoolerSpec {
	specData := CoolerSpec{}
	var socketslist []string

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner .tns-item img", "src")
		specData.Prices = GetPriceLinkFromPangoly(element)

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Release Date":
				specData.ReleaseDate = item.ChildText("td span")
			case "Supported Sockets":
				item.ForEach(".text-left li", func(i int, subitem *colly.HTMLElement) {
					socketslist = append(socketslist, subitem.Text)
				})
				fmt.Println(socketslist)
				specData.Sockets = socketslist
			case "Liquid Cooler":
				// specData.IsLiquidCooler = item.ChildTexts("td")[1]
				specData.IsLiquidCooler = true
			case "Radiator Size":
				specData.LiquidCoolerSize = extractNumberFromString(item.ChildTexts("td")[1])
			case "Noise Level":
				specData.NoiseLevel = item.ChildTexts("td")[1]
			case "Fan RPM":
				specData.FanSpeed = item.ChildTexts("td")[1]
			}
		})
	})
	collector.Visit(link)

	return specData
}

func getCoolerSpecDataFromZol(link string, collector *colly.Collector) CoolerSpec {
	specData := CoolerSpec{
		IsLiquidCooler: false,
		LED:            "NO",
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
			case "散热方式":
				if strings.Contains(convertedData, "水冷") {
					specData.IsLiquidCooler = true
				} else {
					specData.IsLiquidCooler = false
				}
			case "适用范围":
				tempStrList := SplitAny(convertedData, "/")
				specData.Sockets = SocketContainLogic(tempStrList)
			case "发光方式":
				if !strings.Contains(convertedData, "无光") {
					specData.LED = convertedData
				}
			case "风压":
				specData.Pressure = convertedData
			}

			if strings.Contains(convertedHeader, "产品尺") {
				tempStrList := SplitAny(convertedData, "*/")
				heightNum := 0
				for _, testStr := range tempStrList {
					testNum := extractNumberFromString(testStr)
					if testNum > heightNum {
						heightNum = testNum
					}
				}
				specData.AirCoolerHeight = heightNum
				fmt.Println("AirCoolerHeight : ", specData.AirCoolerHeight)
			}

			if strings.Contains(convertedHeader, "水冷排类") {
				specData.LiquidCoolerSize = extractNumberFromString(convertedData)
				fmt.Println("specData.Size : ", specData.LiquidCoolerSize)
			}
			if strings.Contains(convertedData, "dB") {
				specData.NoiseLevel = strings.ReplaceAll(convertedData, "\ufffd\ufffd\u001a", "≤")
			}
			if strings.Contains(convertedData, "RPM") {
				fmt.Println(convertedData)
				if specData.FanSpeed == "" {
					specData.FanSpeed = convertedData
					fmt.Println("specData.FanSpeed: ", convertedData)
				}
			}
			if strings.Contains(convertedData, "CFM") {
				specData.Airflow = convertedData
			}
		})

	})
	collector.Visit(link)
	return specData
}
