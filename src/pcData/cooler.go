package pcData

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type CoolerSpec struct {
	Code           string
	Brand          string
	Name           string
	ReleaseDate    string
	Sockets        []string
	IsLiquidCooler string
	Size           int
	NoiseLevel     string
	FanSpeed       string
	PriceUS        string
	PriceHK        string
	PriceCN        string
	LinkUS         string
	LinkHK         string
	LinkCN         string
	Img            string
}

type CoolerType struct {
	Brand          string
	Name           string
	ReleaseDate    string
	Sockets        []string
	IsLiquidCooler string
	Size           int
	NoiseLevel     string
	FanSpeed       string
	PriceUS        string
	PriceHK        string
	PriceCN        string
	LinkUS         string
	LinkHK         string
	LinkCN         string
	Img            string
}

func GetCoolerSpec(record LinkRecord) CoolerSpec {

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

	cooler := getCoolerSpecData(record.LinkSpec, specCollector)
	cooler.Code = record.Name
	cooler.Brand = record.Brand
	cooler.PriceCN = record.PriceCN
	cooler.PriceHK = ""
	cooler.LinkHK = ""
	cooler.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		cooler.LinkUS = record.LinkUS
	}
	if cooler.Name == "" {
		cooler.Name = record.Name
	}
	return cooler
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

	priceCN := spec.PriceCN
	if priceCN == "" {
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

	return CoolerType{
		Brand:          spec.Brand,
		Name:           spec.Name,
		ReleaseDate:    spec.ReleaseDate,
		Sockets:        spec.Sockets,
		IsLiquidCooler: spec.IsLiquidCooler,
		Size:           spec.Size,
		NoiseLevel:     spec.NoiseLevel,
		FanSpeed:       spec.FanSpeed,
		PriceUS:        priceUS,
		PriceHK:        "",
		PriceCN:        priceCN,
		LinkUS:         spec.LinkUS,
		LinkHK:         spec.LinkHK,
		LinkCN:         spec.LinkCN,
		Img:            tempImg,
	}, isValid
}

func getCoolerSpecData(link string, collector *colly.Collector) CoolerSpec {
	specData := CoolerSpec{}
	var socketslist []string

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner .tns-item img", "src")
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
			case "Supported Sockets":
				item.ForEach(".text-left li", func(i int, subitem *colly.HTMLElement) {
					socketslist = append(socketslist, subitem.Text)
				})
				fmt.Println(socketslist)
				specData.Sockets = socketslist
			case "Liquid Cooler":
				specData.IsLiquidCooler = item.ChildTexts("td")[1]
			case "Radiator Size":
				specData.Size = extractNumberFromString(item.ChildTexts("td")[1])
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
