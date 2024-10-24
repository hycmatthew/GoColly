package pcData

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type CoolerSpec struct {
	Code           string
	Brand          string
	ReleaseDate    string
	Sockets        []string
	IsLiquidCooler string
	Size           string
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
	ReleaseDate    string
	Sockets        []string
	IsLiquidCooler string
	Size           string
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
	cooler.PriceCN = record.PriceCN
	cooler.PriceHK = ""
	cooler.LinkHK = ""
	cooler.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		cooler.LinkUS = record.LinkUS
	}
	return cooler
}

func GetCoolerData(spec CoolerSpec) CoolerType {

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

	return CoolerType{
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
	}
}

func getCoolerSpecData(link string, collector *colly.Collector) CoolerSpec {
	releaseDate := ""
	var socketslist []string
	isLiquidCooler := ""
	size := ""
	noiseLevel := ""
	fanSpeed := ""
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
			case "Supported Sockets":
				item.ForEach("td li", func(i int, subitem *colly.HTMLElement) {
					socketslist = append(socketslist, subitem.Text)
				})
			case "Liquid Cooler":
				isLiquidCooler = item.ChildTexts("td")[1]
			case "Radiator Size":
				size = item.ChildTexts("td")[1]
			case "Noise Level":
				noiseLevel = item.ChildTexts("td")[1]
			case "Fan RPM":
				fanSpeed = item.ChildTexts("td")[1]
			}
		})
	})

	collector.Visit(link)

	return CoolerSpec{
		ReleaseDate:    releaseDate,
		Sockets:        socketslist,
		IsLiquidCooler: isLiquidCooler,
		Size:           size,
		NoiseLevel:     noiseLevel,
		FanSpeed:       fanSpeed,
		PriceUS:        price,
		LinkUS:         usLink,
		Img:            imgLink,
	}
}
