package pcData

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
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

	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"www.newegg.com",
			"newegg.com",
			"www.price.com.hk",
			"price.com.hk",
			"product.pconline.com.cn",
			"pconline.com.cn",
			"pangoly.com",
			"detail.zol.com.cn",
			"zol.com.cn",
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	specCollector := collector.Clone()
	ssdData := SSDSpec{}
	if strings.Contains(record.LinkCN, "zol") {
		ssdData.LinkCN = getDetailsLinkFromZol(record.LinkCN, collector)
	} else {
		if record.LinkSpec != "" {
			ssdData = getSSDSpecData(record.LinkSpec, specCollector)
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
		tempSpec := getSSDSpecDataFromZol(spec.LinkCN, cnCollector)
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
		newSpec.PriceCN = getCNPriceFromPcOnline(spec.LinkCN, cnCollector)

		if newSpec.PriceCN == "" {
			isValid = false
		}
	}

	tempImg := ""
	if strings.Contains(spec.LinkUS, "newegg") {
		newSpec.PriceUS, tempImg = getUSPriceAndImgFromNewEgg(spec.LinkUS, usCollector)

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
		Capacity:    newSpec.Capacity,
		MaxRead:     newSpec.MaxRead,
		MaxWrite:    newSpec.MaxWrite,
		Interface:   newSpec.Interface,
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

func CompareSSDDataLogic(cur SSDType, list []SSDType) SSDType {
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
