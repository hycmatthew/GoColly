package pcData

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type RamSpec struct {
	Code         string
	Brand        string
	Name         string
	Series       string
	Model        string
	Capacity     string
	Type         string
	Speed        int
	Timing       string
	Voltage      string
	Channel      string
	Profile      string
	LED          string
	HeatSpreader bool
	PriceUS      string
	PriceHK      string
	PriceCN      string
	LinkUS       string
	LinkHK       string
	LinkCN       string
	Img          string
}

type RamType struct {
	Brand        string
	Name         string
	Series       string
	Model        string
	Capacity     string
	Type         string
	Speed        int
	Timing       string
	Voltage      string
	Channel      string
	Profile      string
	LED          string
	HeatSpreader bool
	PriceUS      string
	PriceHK      string
	PriceCN      string
	LinkUS       string
	LinkHK       string
	LinkCN       string
	Img          string
}

func GetRamSpec(record LinkRecord) RamSpec {

	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"www.newegg.com",
			"newegg.com",
			"pangoly.com",
			"www.newegg.com",
			"newegg.com",
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	ramData := RamSpec{}

	if record.LinkSpec != "" {
		ramData = getRamSpecData(record.LinkSpec, collector)
	} else {
		ramData = getRamUSPrice(record.LinkUS, collector)
	}
	ramData.Code = record.Name
	ramData.Brand = record.Brand
	ramData.PriceCN = record.PriceCN
	ramData.PriceHK = ""
	ramData.LinkHK = ""
	ramData.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		ramData.LinkUS = record.LinkUS
	}
	if ramData.Name == "" {
		ramData.Name = record.Name
	}
	return ramData
}

func GetRamData(spec RamSpec) RamType {

	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
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
	newSpec := RamSpec{}
	if strings.Contains(spec.LinkUS, "newegg") {
		newSpec = getRamUSPrice(spec.LinkUS, usCollector)
	}

	return RamType{
		Brand:        spec.Brand,
		Name:         spec.Name,
		Series:       newSpec.Series,
		Model:        spec.Model,
		Capacity:     spec.Capacity,
		Speed:        spec.Speed,
		Timing:       spec.Timing,
		Voltage:      spec.Voltage,
		Channel:      newSpec.Channel,
		LED:          spec.LED,
		HeatSpreader: spec.HeatSpreader,
		Profile:      newSpec.Profile,
		PriceUS:      newSpec.PriceUS,
		PriceHK:      spec.PriceHK,
		PriceCN:      priceCN,
		LinkHK:       spec.LinkHK,
		LinkUS:       spec.LinkUS,
		LinkCN:       spec.LinkCN,
		Img:          newSpec.Img,
	}
}

func getRamSpecData(link string, collector *colly.Collector) RamSpec {
	specData := RamSpec{}
	specData.HeatSpreader = false

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
			case "Model":
				specData.Model = item.ChildText("td span")
			case "Speed":
				tempStr := strings.ReplaceAll(item.ChildText("td span"), "-", " ")
				strList := strings.Split(tempStr, " ")
				if strings.Contains(strings.ToUpper(tempStr), "DDR5") {
					specData.Type = "DDR5"
				} else {
					specData.Type = "DDR4"
				}
				if len(strList) > 1 {
					specData.Speed = extractNumberFromString(strList[1])
				}
			case "Size":
				specData.Capacity = item.ChildText("td span")
			case "Timing":
				specData.Timing = item.ChildText("td span")
			case "Voltage":
				specData.Voltage = item.ChildText("td span")
			case "LED Color":
				specData.LED = item.ChildText("td span")
			case "Heat Spreader":
				if strings.ToUpper(item.ChildText("td span")) == "YES" {
					specData.HeatSpreader = true
				}
			}
		})
	})
	collector.Visit(link)

	return specData
}

func getRamUSPrice(link string, collector *colly.Collector) RamSpec {
	specData := RamSpec{}
	specData.HeatSpreader = false

	collectorErrorHandle(collector, link)
	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		specData.PriceUS = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current"))

		element.ForEach(".tab-box .tab-panes tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("th") {
			case "Brand":
				specData.Brand = item.ChildText("td")
			case "Series":
				specData.Series = item.ChildText("td")
			case "Model":
				specData.Model = item.ChildText("td")
			case "Capacity":
				specData.Capacity = item.ChildText("td")
			case "Speed":
				tempStr := strings.ReplaceAll(item.ChildText("td span"), "-", " ")
				strList := strings.Split(tempStr, " ")
				if strings.Contains(strings.ToUpper(tempStr), "DDR5") {
					specData.Type = "DDR5"
				} else {
					specData.Type = "DDR4"
				}
				if len(strList) > 1 {
					specData.Speed = extractNumberFromString(strList[1])
				}
			case "Timing":
				specData.Timing = item.ChildText("td")
			case "Voltage":
				specData.Voltage = item.ChildText("td")
			case "Multi-channel Kit":
				specData.Channel = item.ChildText("td")
			case "BIOS/Performance Profile":
				specData.Profile = item.ChildText("td")
			case "Heat Spreader":
				if strings.ToUpper(item.ChildText("td")) == "YES" {
					specData.HeatSpreader = true
				}
			}
		})
	})

	collector.Visit(link)

	return specData
}
