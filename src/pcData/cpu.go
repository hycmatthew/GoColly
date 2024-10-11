package pcData

import (
	"fmt"
	"net/http"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type LinkRecord struct {
	Brand    string
	Name     string
	PriceCN  string
	LinkSpec string
	LinkCN   string
	LinkUS   string
	LinkHK   string
}

type CPUSpec struct {
	Code            string
	Name            string
	Brand           string
	Socket          string
	Cores           int
	Threads         int
	GPU             string
	SingleCoreScore int
	MultiCoreScore  int
	Power           int
	PriceUS         string
	PriceHK         string
	PriceCN         string
	Img             string
}

type CPUType struct {
	Name            string
	Brand           string
	Socket          string
	Cores           int
	Threads         int
	GPU             string
	SingleCoreScore int
	MultiCoreScore  int
	Power           int
	PriceUS         string
	PriceHK         string
	PriceCN         string
	Img             string
}

func GetCPUSpec(record LinkRecord) CPUSpec {
	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"nanoreview.net",
			"www.newegg.com",
			"newegg.com",
		),
		colly.AllowURLRevisit(),
	)

	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	cpuData := getCPUSpecData(record.LinkSpec, collector)
	cpuData.Code = record.Name
	cpuData.PriceCN = record.LinkCN
	cpuData.PriceUS = record.LinkUS
	cpuData.PriceHK = record.LinkHK
	// cpuData.PriceHK = getCPUHKPrice(hkLink, hkCollector)
	return cpuData
}

func GetCPUData(spec CPUSpec) CPUType {

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

	priceCN := getCPUCNPrice(spec.PriceCN, cnCollector)
	PriceUS, tempImg := getCPUUSPrice(spec.PriceCN, usCollector)

	return CPUType{
		Name:            spec.Name,
		Brand:           spec.Brand,
		Cores:           spec.Cores,
		Threads:         spec.Threads,
		Socket:          spec.Socket,
		GPU:             spec.GPU,
		SingleCoreScore: spec.SingleCoreScore,
		MultiCoreScore:  spec.MultiCoreScore,
		Power:           spec.Power,
		PriceCN:         priceCN,
		PriceUS:         PriceUS,
		PriceHK:         "",
		Img:             tempImg,
	}
}

func getCPUSpecData(link string, collector *colly.Collector) CPUSpec {
	name := ""
	brand := ""
	socket := ""
	cores := 0
	thread := 0
	tdp := 0
	gpu := ""
	singleCoreScore := 0
	muitiCoreScore := 0

	collectorErrorHandle(collector, link)

	collector.OnHTML("#the-app", func(element *colly.HTMLElement) {

		element.ForEach(".two-columns-item .score-bar", func(i int, item *colly.HTMLElement) {
			switch item.ChildText(".score-bar-name") {
			case "Cinebench R23 (Single-Core)":
				singleCoreScore = extractNumberFromString(item.ChildText(".score-bar-result-number"))
			case "Cinebench R23 (Multi-Core)":
				muitiCoreScore = extractNumberFromString(item.ChildText(".score-bar-result-number"))
			}
		})

		element.ForEach(".specs-table tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText(".cell-h") {
			case "Vendor":
				brand = item.ChildText(".cell-s")
			case "Total Cores":
				cores = extractNumberFromString(item.ChildText(".cell-s"))
			case "Total Threads":
				thread = extractNumberFromString(item.ChildText(".cell-s"))
			case "Socket":
				socket = item.ChildText("td")
			case "Integrated GPU":
				gpu = item.ChildText("td")
			case "TDP (PL1)":
				tdp = extractNumberFromString(item.ChildText("td"))
			case "Max. Boost TDP (PL2)":
				tempTdp := extractNumberFromString(item.ChildText("td"))
				if tempTdp > tdp {
					tdp = tempTdp
				}
			}
		})

		name = element.ChildText(".card-head .title-h1")
	})

	collector.Visit(link)

	return CPUSpec{
		Name:            name,
		Brand:           brand,
		Cores:           cores,
		Threads:         thread,
		Socket:          socket,
		GPU:             gpu,
		SingleCoreScore: singleCoreScore,
		MultiCoreScore:  muitiCoreScore,
		Power:           tdp,
	}
}

func getCPUUSPrice(link string, collector *colly.Collector) (string, string) {
	imgLink, price := "", ""

	collectorErrorHandle(collector, link)
	fmt.Println(collector.AllowedDomains)

	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		price = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current"))
	})
	collector.Visit(link)
	return price, imgLink
}

func getCPUHKPrice(link string, collector *colly.Collector) string {
	price := ""
	collectorErrorHandle(collector, link)

	collector.OnHTML(".line-05", func(element *colly.HTMLElement) {

		element.ForEach(".product-price", func(i int, item *colly.HTMLElement) {
			if price == "" {
				price = extractFloatStringFromString(element.ChildText("span"))
			}
		})
	})

	collector.Visit(link)
	return price
}

func getCPUCNPrice(link string, collector *colly.Collector) string {
	price := ""

	collectorErrorHandle(collector, link)

	collector.OnHTML(".product-mallSales", func(element *colly.HTMLElement) {
		price = extractFloatStringFromString(element.ChildText("em.price"))
	})

	collector.Visit(link)
	return price
}

func collectorErrorHandle(collector *colly.Collector, link string) {
	collector.OnRequest(func(r *colly.Request) {
		// USER_AGENT = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36'
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
	})

	collector.OnError(func(response *colly.Response, err error) {
		fmt.Println("请求期间发生错误,则调用:", err, " - link: ", link)
	})

	collector.OnResponse(func(response *colly.Response) {
		fmt.Println("收到响应后调用:", response.Request.URL)
	})
}
