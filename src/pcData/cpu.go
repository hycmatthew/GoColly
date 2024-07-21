package pcData

import (
	"fmt"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type CPURecord struct {
	Name     string
	LinkSpec string
	LinkCN   string
	LinkHK   string
	LinkUS   string
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

func GetCPUData(specLink string, enLink string, cnLink string, hkLink string) []CPUType {

	collector := colly.NewCollector(
		colly.UserAgent(req.DefaultClient().ImpersonateChrome().Headers.Get("user-agent")),
		colly.AllowedDomains(
			// "https://nanoreview.net",
			"nanoreview.net",
			"www.newegg.com",
			"newegg.com",
			// "https://cu.manmanbuy.com",
			"cu.manmanbuy.com",
			"www.price.com.hk",
			"price.com.hk",
		),
		colly.AllowURLRevisit(),
	)

	usCollector := collector.Clone()
	cnCollector := collector.Clone()
	hkCollector := collector.Clone()

	var cpuList []CPUType
	cpuData := getSpecData(specLink, collector)
	cpuData.PriceUS = getUSPrice(enLink, usCollector)
	cpuData.PriceCN = getCNPrice(cnLink, cnCollector)
	cpuData.PriceHK = getHKPrice(hkLink, hkCollector)

	fmt.Println(cpuData)

	cpuList = append(cpuList, cpuData)

	return cpuList
}

func getSpecData(link string, collector *colly.Collector) CPUType {
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
		/*
			fmt.Println("record logic!!")
			fmt.Println(brand)
			fmt.Println(cores)
			fmt.Println(thread)
			fmt.Println(socket)
			fmt.Println(singleCoreScore)
			fmt.Println(muitiCoreScore)
			fmt.Println(gpu)
			fmt.Println(tdp)
		*/
	})

	collector.Visit(link)

	return CPUType{
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

func getUSPrice(link string, collector *colly.Collector) float64 {
	price := 0.0

	collectorErrorHandle(collector, link)
	fmt.Println(collector.AllowedDomains)

	collector.OnHTML(".row-side .product-buy-box", func(element *colly.HTMLElement) {
		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText("li.price-current")), 64); err == nil {
			price = s
			fmt.Println(price)
		}
	})

	collector.Visit(link)
	return price
}

func getHKPrice(link string, collector *colly.Collector) float64 {
	price := 0.0

	collectorErrorHandle(collector, link)

	collector.OnHTML(".line-05", func(element *colly.HTMLElement) {

		element.ForEach(".product-price", func(i int, item *colly.HTMLElement) {
			fmt.Println(extractFloatStringFromString(element.ChildText("span")))
			if price == 0.0 {
				if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText("span")), 64); err == nil {
					price = s
					fmt.Println(price)
				} else {
					fmt.Println(err)
				}
			}
		})
	})

	collector.Visit(link)
	return price
}

func getCNPrice(link string, collector *colly.Collector) float64 {
	price := 0.0

	collectorErrorHandle(collector, link)

	collector.OnHTML(".articlehead .t", func(element *colly.HTMLElement) {
		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText("span")), 64); err == nil {
			price = s
			fmt.Println(price)
		} else {
			fmt.Println(err)
		}
	})

	collector.Visit("https://cu.manmanbuy.com/discuxiao_8908724.aspx")
	return price
}

func collectorErrorHandle(collector *colly.Collector, link string) {
	collector.OnRequest(func(r *colly.Request) {
		// USER_AGENT = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36'

		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36")
	})

	collector.OnError(func(response *colly.Response, err error) {
		fmt.Println("请求期间发生错误,则调用:", err, " - link: ", link)
	})

	collector.OnResponse(func(response *colly.Response) {
		fmt.Println("收到响应后调用:", response.Request.URL)
	})
}

func linkSetupLogic(link string) string {
	return ("https://" + link)
}
