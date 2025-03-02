package pcData

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
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
	PriceUS          string
	PriceHK          string
	PriceCN          string
	LinkUS           string
	LinkHK           string
	LinkCN           string
	Img              string
}

type CoolerType struct {
	Id               string
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
	PriceUS          string
	PriceHK          string
	PriceCN          string
	LinkUS           string
	LinkHK           string
	LinkCN           string
	Img              string
}

func GetCoolerSpec(record LinkRecord) CoolerSpec {
	coolerData := CoolerSpec{
		LinkCN: record.LinkCN,
	}

	if strings.Contains(record.LinkCN, "zol") {
		coolerData.LinkCN = getDetailsLinkFromZol(record.LinkCN, CreateCollector())
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
	coolerData.PriceCN = record.PriceCN
	coolerData.PriceHK = ""
	coolerData.LinkHK = ""
	if record.LinkUS != "" {
		coolerData.LinkUS = record.LinkUS
	}
	if coolerData.Name == "" {
		coolerData.Name = record.Name
	}
	coolerData.Name = RemoveBrandsFromName(coolerData.Brand, coolerData.Name)
	return coolerData
}

func GetCoolerData(spec CoolerSpec) (CoolerType, bool) {

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
		tempSpec := getCoolerSpecDataFromZol(spec.LinkCN, cnCollector)

		newSpec.Img = tempSpec.Img
		if tempSpec.PriceCN != "" {
			newSpec.PriceCN = tempSpec.PriceCN
		}
		newSpec.IsLiquidCooler = tempSpec.IsLiquidCooler
		newSpec.Sockets = tempSpec.Sockets
		newSpec.AirCoolerHeight = tempSpec.AirCoolerHeight
		newSpec.LiquidCoolerSize = tempSpec.LiquidCoolerSize
		newSpec.NoiseLevel = tempSpec.NoiseLevel
		newSpec.FanSpeed = tempSpec.FanSpeed
		newSpec.Airflow = tempSpec.Airflow
		newSpec.Pressure = tempSpec.Pressure
		newSpec.LED = tempSpec.LED
		if strings.Contains(newSpec.Name, "RGB") {
			newSpec.LED = "RGB"
		}
		if strings.Contains(newSpec.Name, "ARGB") {
			newSpec.LED = "ARGB"
		}
		if newSpec.PriceCN == "" {
			isValid = false
		}
	}

	if newSpec.PriceCN == "" && strings.Contains(spec.LinkCN, "pconline") {
		newSpec.PriceCN = getCNPriceFromPcOnline(spec.LinkCN, cnCollector)

		if newSpec.PriceCN == "" {
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

	return CoolerType{
		Id:               SetProductId(spec.Brand, spec.Code),
		Brand:            spec.Brand,
		Name:             spec.Name,
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
		PriceUS:          priceUS,
		PriceHK:          "",
		PriceCN:          newSpec.PriceCN,
		LinkUS:           spec.LinkUS,
		LinkHK:           spec.LinkHK,
		LinkCN:           spec.LinkCN,
		Img:              tempImg,
	}, isValid
}

func getCoolerSpecData(link string, collector *colly.Collector) CoolerSpec {
	specData := CoolerSpec{}
	var socketslist []string

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner .tns-item img", "src")
		specData.PriceUS, specData.LinkUS = GetPriceLinkFromPangoly(element)

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

		mallPrice := extractFloatStringFromString(element.ChildText("side .goods-card .item-b2cprice span"))
		// otherPrice := extractFloatStringFromString(element.ChildText(".price__merchant .price"))
		normalPrice := extractFloatStringFromString(element.ChildText(".side .goods-card .goods-card__price span"))
		if mallPrice != "" {
			specData.PriceCN = mallPrice
			specData.LinkCN = element.ChildAttr("side .goods-card .item-b2cprice span a", "href")
		} else {
			specData.PriceCN = normalPrice
		}

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
