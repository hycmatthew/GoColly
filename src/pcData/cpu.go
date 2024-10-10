package pcData

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

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
	PriceUS         float64
	PriceHK         float64
	PriceCN         float64
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

	cpuData := getCPUSpecData(record.LinkSpec, collector)
	cpuData.Code = record.Name
	cpuData.PriceCN = record.LinkCN
	cpuData.PriceUS = record.LinkUS
	cpuData.PriceHK = record.LinkHK
	// cpuData.PriceHK = getCPUHKPrice(hkLink, hkCollector)
	return cpuData
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

func getCPUUSPrice(link string, collector *colly.Collector) (float64, string) {
	imgLink := ""
	price := 0.0

	collectorErrorHandle(collector, link)
	fmt.Println(collector.AllowedDomains)

	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")

		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current")), 64); err == nil {
			price = s
			//fmt.Println(price)
		}
	})

	collector.Visit(link)
	return price, imgLink
}

func getCPUHKPrice(link string, collector *colly.Collector) float64 {
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

func getCPUCNPrice(link string, collector *colly.Collector) float64 {
	price := 0.0

	collectorErrorHandle(collector, link)

	collector.OnHTML(".product-mallSales", func(element *colly.HTMLElement) {
		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText("em.price")), 64); err == nil {
			price = s
			// fmt.Println(price)
		} else {
			fmt.Println(err)
		}
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
