package pcData

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type CaseSpec struct {
	Code          string
	Brand         string
	Name          string
	ReleaseDate   string
	Color         string
	CaseSize      string
	PowerSupply   bool
	DriveBays2    int
	DriveBays3    int
	Compatibility string
	Dimensions    []int
	MaxVGAlength  int
	SlotsNum      int
	PriceUS       string
	PriceHK       string
	PriceCN       string
	LinkUS        string
	LinkHK        string
	LinkCN        string
	Img           string
}

type CaseType struct {
	Brand         string
	Name          string
	ReleaseDate   string
	Color         string
	CaseSize      string
	PowerSupply   bool
	DriveBays2    int
	DriveBays3    int
	Compatibility string
	Dimensions    []int
	MaxVGAlength  int
	SlotsNum      int
	PriceUS       string
	PriceHK       string
	PriceCN       string
	LinkUS        string
	LinkHK        string
	LinkCN        string
	Img           string
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
	priceUS, tempImg := spec.PriceUS, spec.Img
	if strings.Contains(spec.LinkUS, "newegg") {
		priceUS, tempImg = getUSPriceAndImgFromNewEgg(spec.LinkUS, usCollector)
	}

	return CaseType{
		Brand:         spec.Brand,
		Name:          spec.Name,
		ReleaseDate:   spec.ReleaseDate,
		CaseSize:      spec.CaseSize,
		Color:         spec.Color,
		PowerSupply:   spec.PowerSupply,
		DriveBays2:    spec.DriveBays2,
		DriveBays3:    spec.DriveBays3,
		Compatibility: spec.Compatibility,
		Dimensions:    spec.Dimensions,
		MaxVGAlength:  spec.MaxVGAlength,
		PriceUS:       priceUS,
		PriceHK:       "",
		PriceCN:       priceCN,
		LinkUS:        spec.LinkUS,
		LinkHK:        spec.LinkHK,
		LinkCN:        spec.LinkCN,
		Img:           tempImg,
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
