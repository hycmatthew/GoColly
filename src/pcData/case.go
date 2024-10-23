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
	ReleaseDate   string
	CaseSize      string
	PowerSupply   bool
	DriveBays2    int
	DriveBays3    int
	Compatibility string
	Dimensions    string
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
	ReleaseDate   string
	CaseSize      string
	PowerSupply   bool
	DriveBays2    int
	DriveBays3    int
	Compatibility string
	Dimensions    string
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
	caseData.Code = record.Name
	caseData.PriceCN = record.PriceCN
	caseData.PriceHK = ""
	caseData.LinkHK = ""
	caseData.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		caseData.LinkUS = record.LinkUS
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
	if priceCN != "" {
		priceCN = getCNPriceFromPcOnline(spec.LinkCN, cnCollector)
	}
	priceUS, tempImg := spec.PriceUS, spec.Img
	if strings.Contains(spec.LinkUS, "newegg") {
		priceUS, tempImg = getUSPriceAndImgFromNewEgg(spec.LinkUS, usCollector)
	}

	return CaseType{
		Brand:         spec.Brand,
		ReleaseDate:   spec.ReleaseDate,
		CaseSize:      spec.CaseSize,
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
	releaseDate := ""
	caseSize := ""
	powerSupply := false
	driveBays2 := 0
	driveBays3 := 0
	compatibility := ""
	dimensions := ""
	maxVGAlength := 0
	slots := 0
	price := ""
	usLink := ""
	imgLink := ""

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".tns-inner .tns-item img", "src")
		loopBreak := false

		element.ForEach("table.table-prices tr", func(i int, item *colly.HTMLElement) {
			if !loopBreak {
				price = extractFloatStringFromString(item.ChildText(".detail-purchase strong"))
				tempLink := item.ChildAttr(".detail-purchase", "href")

				if strings.Contains(tempLink, "amazon") {
					amazonLink := strings.Split(tempLink, "?tag=")[0]
					usLink = amazonLink
					loopBreak = true
				}
				if strings.Contains(tempLink, "newegg") {
					neweggLink := strings.Split(tempLink, "url=")[1]
					UnescapeLink, _ := url.QueryUnescape(neweggLink)
					usLink = strings.Split(UnescapeLink, "\u0026")[0]
					loopBreak = true
				}
			}
		})

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Release Date":
				releaseDate = item.ChildText("td span")
			case "Type":
				caseSize = item.ChildTexts("td")[1]
			case "Includes Power Supply":
				if item.ChildTexts("td")[1] != "No" {
					powerSupply = true
				}
			case `Internal 2.5" Drive Bays`:
				driveBays2 = extractNumberFromString(item.ChildTexts("td")[1])
			case `Internal 3.5" Drive Bays`:
				driveBays3 = extractNumberFromString(item.ChildTexts("td")[1])
			case "Motherboard Compatibility":
				compatibility = item.ChildTexts("td")[1]
			case "Dimensions":
				dimensions = item.ChildTexts("td")[1]
			case "Max VGA length allowance":
				maxVGAlength = extractNumberFromString(item.ChildTexts("td")[1])
			case "Expansion Slots":
				slots = extractNumberFromString(item.ChildTexts("td")[1])
			}
		})
	})

	collector.Visit(link)

	return CaseSpec{
		ReleaseDate:   releaseDate,
		CaseSize:      caseSize,
		PowerSupply:   powerSupply,
		DriveBays2:    driveBays2,
		DriveBays3:    driveBays3,
		Compatibility: compatibility,
		Dimensions:    dimensions,
		MaxVGAlength:  maxVGAlength,
		SlotsNum:      slots,
		PriceUS:       price,
		LinkUS:        usLink,
		Img:           imgLink,
	}
}
