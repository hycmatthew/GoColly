package pcData

import (
	"net/http"
	"net/url"
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
	Brand       string
	Name        string
	ReleaseDate string
	Model       string
	Capacity    string
	MaxRead     int
	MaxWrite    int
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
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	specCollector := collector.Clone()
	ssdData := SSDSpec{}
	if record.LinkSpec != "" {
		ssdData = getSSDSpecData(record.LinkSpec, specCollector)
	}
	ssdData.Code = record.Name
	ssdData.Brand = record.Brand
	ssdData.PriceCN = record.PriceCN
	ssdData.PriceHK = ""
	ssdData.LinkHK = ""
	ssdData.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		ssdData.LinkUS = record.LinkUS
	}
	if ssdData.Name == "" {
		ssdData.Name = record.Name
	}
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

	priceCN := spec.PriceCN
	if priceCN == "" {
		if spec.Brand == "zhitai" {
			tempSpec := getZhiTaiDataFromPcOnline(spec.LinkCN, cnCollector)
			codeStringList := strings.Split(spec.Code, " ")
			spec.Capacity = codeStringList[len(codeStringList)-1]
			spec.FormFactor = tempSpec.FormFactor
			spec.FlashType = tempSpec.FlashType
			spec.Interface = tempSpec.Interface
			spec.MaxRead = tempSpec.MaxRead
			spec.MaxWrite = tempSpec.MaxWrite
			priceCN = tempSpec.PriceCN
		} else {
			priceCN = getCNPriceFromPcOnline(spec.LinkCN, cnCollector)
		}

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

	return SSDType{
		Brand:       spec.Brand,
		Name:        spec.Name,
		ReleaseDate: spec.ReleaseDate,
		Model:       spec.Model,
		Capacity:    spec.Capacity,
		MaxRead:     spec.MaxRead,
		MaxWrite:    spec.MaxWrite,
		Interface:   spec.Interface,
		FlashType:   spec.FlashType,
		FormFactor:  spec.FormFactor,
		PriceUS:     priceUS,
		PriceHK:     "",
		PriceCN:     priceCN,
		LinkUS:      spec.LinkUS,
		LinkHK:      spec.LinkHK,
		LinkCN:      spec.LinkCN,
		Img:         tempImg,
	}, isValid
}

func getSSDSpecData(link string, collector *colly.Collector) SSDSpec {
	specData := SSDSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {

		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner img", "src")
		loopBreak := false

		element.ForEach("table.table-prices tr", func(i int, item *colly.HTMLElement) {
			if !loopBreak {
				specData.PriceUS = extractFloatStringFromString(item.ChildText(".detail-purchase strong"))
				tempLink := item.ChildAttr(".detail-purchase", "href")

				if strings.Contains(tempLink, "amazon") && specData.LinkUS == "" {
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
