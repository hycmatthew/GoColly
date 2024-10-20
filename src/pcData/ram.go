package pcData

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chromedp/chromedp"
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
	ramData.PriceCN = record.LinkCN
	return ramData
}

func GetRamData(spec RamSpec) RamType {

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

	ramData := getRamUSPrice(spec.LinkUS, collector)

	return RamType{
		Brand:    ramData.Brand,
		Series:   ramData.Series,
		Model:    ramData.Model,
		Capacity: ramData.Capacity,
		Speed:    ramData.Speed,
		Timing:   ramData.Timing,
		Voltage:  ramData.Voltage,
		Channel:  ramData.Channel,
		Profile:  ramData.Profile,
		PriceUS:  ramData.PriceUS,
		PriceHK:  ramData.PriceHK,
		PriceCN:  ramData.PriceCN,
		LinkHK:   ramData.LinkHK,
		LinkUS:   ramData.LinkUS,
		LinkCN:   ramData.LinkCN,
		Img:      ramData.Img,
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

func getRamHKPrice(link string, collector *colly.Collector) float64 {
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

func getRamCNPrice(link string) string {
	fmt.Println(link)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		allocCtx,
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	// navigate to a page, wait for an element, click
	var cnPrice string
	err := chromedp.Run(ctx,
		chromedp.Navigate(link),
		// wait for footer element is visible (ie, page is loaded)
		chromedp.Sleep(600*time.Second),
		// retrieve the value of the textarea
		chromedp.Value(`.p-price .price`, &cnPrice),
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cnPrice)
	return cnPrice
}
