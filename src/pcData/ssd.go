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

	ssdData := getSSDSpecData(record.LinkSpec, specCollector)
	ssdData.Code = record.Name
	ssdData.PriceCN = record.PriceCN
	ssdData.PriceHK = ""
	ssdData.LinkHK = ""
	ssdData.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		ssdData.LinkUS = record.LinkUS
	}
	return ssdData
}

func GetSSDData(spec SSDSpec) SSDType {

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

	return SSDType{
		Brand:       spec.Brand,
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
	}
}

func getSSDSpecData(link string, collector *colly.Collector) SSDSpec {
	releaseDate := ""
	model := ""
	capacity := ""
	maxRead := 0
	maxWrite := 0
	ssdInterface := ""
	flashType := ""
	formFactor := ""
	price := ""
	usLink := ""
	imgLink := ""

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".tns-inner img", "src")
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
			case "Model":
				model = strings.Split(item.ChildTexts("td")[1], "\n")[0]
			case "Release Date":
				releaseDate = item.ChildText("td span")
			case "Capacity":
				capacity = item.ChildTexts("td")[1]
			case "Interface":
				ssdInterface = item.ChildTexts("td")[1]
			case "Form Factor":
				formFactor = item.ChildTexts("td")[1]
			case "NAND Flash Type":
				flashType = item.ChildTexts("td")[1]
			case "Max Sequential Read":
				maxRead = extractNumberFromString(item.ChildTexts("td")[1])
			case "Max Sequential Write":
				maxWrite = extractNumberFromString(item.ChildTexts("td")[1])
			}
		})
	})

	collector.Visit(link)

	return SSDSpec{
		ReleaseDate: releaseDate,
		Model:       model,
		Capacity:    capacity,
		MaxRead:     maxRead,
		MaxWrite:    maxWrite,
		Interface:   ssdInterface,
		FlashType:   flashType,
		FormFactor:  formFactor,
		PriceUS:     price,
		LinkUS:      usLink,
		Img:         imgLink,
	}
}
