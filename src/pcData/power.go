package pcData

import (
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

	ssdData := getPowerSpecData(record.LinkSpec, specCollector)
	ssdData.Code = record.Name
	ssdData.Brand = record.Brand
	ssdData.PriceCN = record.PriceCN
	ssdData.PriceHK = ""
	ssdData.LinkHK = ""
	ssdData.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		ssdData.LinkUS = record.LinkUS
	}
	return ssdData
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
	if priceCN != "" {
		priceCN = getCNPriceFromPcOnline(spec.LinkCN, cnCollector)

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
