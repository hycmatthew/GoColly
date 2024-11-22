package pcData

import (
	"net/http"
	"strings"

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
	LinkUS          string
	LinkHK          string
	LinkCN          string
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
	LinkUS          string
	LinkHK          string
	LinkCN          string
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
	cpuData.PriceCN = record.PriceCN
	cpuData.PriceHK = ""
	cpuData.LinkHK = ""
	cpuData.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		cpuData.LinkUS = record.LinkUS
	}
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

	priceCN := spec.PriceCN
	if priceCN == "" {
		priceCN = getCNPriceFromPcOnline(spec.LinkCN, cnCollector)
	}

	priceUS, tempImg := spec.PriceUS, spec.Img
	if strings.Contains(spec.LinkUS, "newegg") {
		priceUS, tempImg = getUSPriceAndImgFromNewEgg(spec.LinkUS, usCollector)
	}

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
		LinkUS:          spec.LinkUS,
		LinkHK:          "",
		LinkCN:          spec.LinkCN,
		PriceCN:         priceCN,
		PriceUS:         priceUS,
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
				socket = item.ChildText(".cell-s")
			case "Integrated GPU":
				gpu = item.ChildText(".cell-s")
			case "TDP (PL1)":
				tdp = extractNumberFromString(item.ChildText(".cell-s"))
			case "Max. Boost TDP (PL2)":
				tempTdp := extractNumberFromString(item.ChildText(".cell-s"))
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
		Socket:          strings.Replace(socket, "-", "", -1),
		GPU:             gpu,
		SingleCoreScore: singleCoreScore,
		MultiCoreScore:  muitiCoreScore,
		Power:           tdp,
	}
}
