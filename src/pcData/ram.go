package pcData

import (
	"net/http"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type RamSpec struct {
	Code     string
	Brand    string
	Series   string
	Model    string
	Capacity string
	Speed    string
	Timing   string
	Voltage  string
	Channel  string
	Profile  string
	PriceUS  string
	PriceHK  string
	PriceCN  string
	LinkUS   string
	LinkHK   string
	LinkCN   string
	Img      string
}

type RamType struct {
	Brand    string
	Series   string
	Model    string
	Capacity string
	Speed    string
	Timing   string
	Voltage  string
	Channel  string
	Profile  string
	PriceUS  string
	PriceHK  string
	PriceCN  string
	LinkUS   string
	LinkHK   string
	LinkCN   string
	Img      string
}

func GetRamSpec(record LinkRecord) RamSpec {

	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"www.newegg.com",
			"newegg.com",
			"pangoly.com",
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	ramData := getRamUSPrice(record.LinkUS, collector)
	ramData.Code = record.Name
	ramData.PriceCN = record.PriceCN
	ramData.PriceHK = ""
	ramData.LinkHK = ""
	ramData.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		ramData.LinkUS = record.LinkUS
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
	priceUS, tempImg := spec.PriceUS, spec.Img
	if strings.Contains(spec.LinkUS, "newegg") {
		priceUS, tempImg = getUSPriceAndImgFromNewEgg(spec.LinkUS, usCollector)
	}

	return RamType{
		Brand:    spec.Brand,
		Series:   spec.Series,
		Model:    spec.Model,
		Capacity: spec.Capacity,
		Speed:    spec.Speed,
		Timing:   spec.Timing,
		Voltage:  spec.Voltage,
		Channel:  spec.Channel,
		Profile:  spec.Profile,
		PriceUS:  priceUS,
		PriceHK:  spec.PriceHK,
		PriceCN:  priceCN,
		LinkHK:   spec.LinkHK,
		LinkUS:   spec.LinkUS,
		LinkCN:   spec.LinkCN,
		Img:      tempImg,
	}
}

func getRamUSPrice(link string, collector *colly.Collector) RamSpec {
	brand := ""
	series := ""
	model := ""
	capacity := ""
	speed := ""
	timing := ""
	voltage := ""
	channel := ""
	profile := ""
	price := ""
	imgLink := ""

	collectorErrorHandle(collector, link)
	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		price = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current"))

		element.ForEach(".tab-box .tab-panes tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("th") {
			case "Brand":
				brand = item.ChildText("td")
			case "Series":
				series = item.ChildText("td")
			case "Model":
				model = item.ChildText("td")
			case "Capacity":
				capacity = item.ChildText("td")
			case "Speed":
				speed = item.ChildText("td")
			case "Timing":
				timing = item.ChildText("td")
			case "Voltage":
				voltage = item.ChildText("td")
			case "Multi-channel Kit":
				channel = item.ChildText("td")
			case "BIOS/Performance Profile":
				profile = item.ChildText("td")
			}
		})
	})

	collector.Visit(link)

	return RamSpec{
		Brand:    brand,
		Series:   series,
		Model:    model,
		Capacity: capacity,
		Speed:    speed,
		Timing:   timing,
		Voltage:  voltage,
		Channel:  channel,
		Profile:  profile,
		PriceUS:  price,
		PriceHK:  "",
		PriceCN:  "",
		Img:      imgLink,
	}
}
