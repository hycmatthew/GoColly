package pcData

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	ssdData.Brand = record.Brand
	ssdData.LinkCN = record.LinkCN
	if ssdData.LinkUS == "" {
		ssdData.LinkUS = record.LinkUS
	}
	ssdData.LinkHK = record.LinkHK
	if record.PriceCN != "" {
		ssdData.PriceCN = record.PriceCN
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

	priceCN := getPowerCNPrice(spec.PriceCN, cnCollector)
	priceUS, tempImg := getPowerUSPrice(spec.PriceCN, usCollector)

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

func getPowerUSPrice(link string, collector *colly.Collector) (string, string) {
	price, imgLink := "", ""

	collectorErrorHandle(collector, link)
	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		price = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current"))
	})

	collector.Visit(link)
	return price, imgLink
}

func getPowerHKPrice(link string, collector *colly.Collector) float64 {
	price := 0.0

	collectorErrorHandle(collector, link)

	collector.OnHTML(".line-05", func(element *colly.HTMLElement) {

		element.ForEach(".product-price", func(i int, item *colly.HTMLElement) {
			fmt.Println(extractFloatStringFromString(element.ChildText("span")))
			if price == 0.0 {
				if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText("span")), 64); err == nil {
					price = s
					//fmt.Println(price)
				} else {
					fmt.Println(err)
				}
			}
		})
	})

	collector.Visit(link)
	return price
}

func getPowerCNPrice(link string, collector *colly.Collector) string {
	price := ""
	collectorErrorHandle(collector, link)

	collector.OnHTML(".product-mallSales", func(element *colly.HTMLElement) {
		price = extractFloatStringFromString(element.ChildText("em.price"))
	})

	collector.Visit(link)
	return price
}
