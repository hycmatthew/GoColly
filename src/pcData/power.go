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
	Model       string
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
	Model       string
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
	ssdData.PriceCN = record.PriceCN
	ssdData.PriceHK = ""
	ssdData.LinkHK = ""
	ssdData.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		ssdData.LinkUS = record.LinkUS
	}
	return ssdData
}

func GetPowerData(spec PowerSpec) PowerType {

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

	return PowerType{
		Brand:       spec.Brand,
		Model:       spec.Model,
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
	}
}

func getPowerSpecData(link string, collector *colly.Collector) PowerSpec {
	model := ""
	releaseDate := ""
	wattage := 0
	size := ""
	modular := ""
	efficiency := ""
	length := 0
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
				model = item.ChildTexts("td")[1]
			case "Release Date":
				releaseDate = item.ChildText("td span")
			case "Wattage":
				wattage = extractNumberFromString(item.ChildTexts("td")[1])
			case "Type":
				size = item.ChildTexts("td")[1]
			case "Modular":
				modular = item.ChildTexts("td")[1]
			case "Efficiency":
				efficiency = item.ChildTexts("td")[1]
			case "Length":
				length = extractNumberFromString(item.ChildTexts("td")[1])
			}
		})
	})

	collector.Visit(link)

	return PowerSpec{
		ReleaseDate: releaseDate,
		Model:       model,
		Wattage:     wattage,
		Size:        size,
		Modular:     modular,
		Efficiency:  efficiency,
		Length:      length,
		PriceUS:     price,
		LinkUS:      usLink,
		Img:         imgLink,
	}
}
