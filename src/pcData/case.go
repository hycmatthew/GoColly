package pcData

import (
	"fmt"
	"net/http"
	"net/url"
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
	Compatibility      string
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
	Brand              string
	Name               string
	ReleaseDate        string
	Color              string
	CaseSize           string
	PowerSupply        bool
	DriveBays2         int
	DriveBays3         int
	Compatibility      string
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

	caseData := getCaseSpecData(record.LinkSpec, specCollector)
	caseData.Brand = record.Brand
	caseData.Code = record.Name
	caseData.PriceCN = record.PriceCN
	caseData.PriceHK = ""
	caseData.LinkHK = ""
	caseData.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		caseData.LinkUS = record.LinkUS
	}
	if caseData.Name == "" {
		caseData.Name = record.Name
	}
	return caseData
}

func GetCaseData(spec CaseSpec) CaseType {

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

	priceCN := spec.PriceCN
	if priceCN == "" {
		priceCN = getCNPriceFromPcOnline(spec.LinkCN, cnCollector)
	}

	newSpec := CaseSpec{}
	if strings.Contains(spec.LinkUS, "newegg") {
		newSpec = getCaseUSPrice(spec.LinkUS, usCollector)
	}

	return CaseType{
		Brand:              spec.Brand,
		Name:               spec.Name,
		ReleaseDate:        spec.ReleaseDate,
		CaseSize:           spec.CaseSize,
		Color:              spec.Color,
		PowerSupply:        spec.PowerSupply,
		DriveBays2:         spec.DriveBays2,
		DriveBays3:         spec.DriveBays3,
		Compatibility:      spec.Compatibility,
		Dimensions:         spec.Dimensions,
		MaxVGAlength:       spec.MaxVGAlength,
		RadiatorSupport:    newSpec.RadiatorSupport,
		MaxCpuCoolorHeight: newSpec.MaxCpuCoolorHeight,
		SlotsNum:           spec.SlotsNum,
		PriceUS:            newSpec.PriceUS,
		PriceHK:            "",
		PriceCN:            priceCN,
		LinkUS:             spec.LinkUS,
		LinkHK:             spec.LinkHK,
		LinkCN:             spec.LinkCN,
		Img:                newSpec.Img,
	}
}

func getCaseSpecData(link string, collector *colly.Collector) CaseSpec {
	specData := CaseSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner .tns-item img", "src")
		loopBreak := false

		element.ForEach("table.table-prices tr", func(i int, item *colly.HTMLElement) {
			if !loopBreak {
				specData.PriceUS = extractFloatStringFromString(item.ChildText(".detail-purchase strong"))
				tempLink := item.ChildAttr(".detail-purchase", "href")

				if strings.Contains(tempLink, "amazon") && specData.LinkUS == "" {
					amazonLink := strings.Split(tempLink, "?tag=")[0]
					specData.LinkUS = amazonLink
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
				specData.Compatibility = item.ChildTexts("td")[1]
			case "Dimensions":
				tempDimensions := strings.Split(item.ChildTexts("td")[1], "x")
				var dimensionsList []int
				for _, item := range tempDimensions {
					dimensionsList = append(dimensionsList, extractNumberFromString(item))
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
		specData.PriceUS = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current"))

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
