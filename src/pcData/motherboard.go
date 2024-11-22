package pcData

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type MotherboardSpec struct {
	Code       string
	Name       string
	Brand      string
	Socket     string
	Chipset    string
	RamSlot    int
	RamType    string
	RamSupport []string
	RamMax     int
	Pcie16Slot int
	Pcie4Slot  int
	Pcie1Slot  int
	M2Slot     int
	SataSlot   int
	FormFactor string
	Wireless   bool
	PriceUS    string
	PriceHK    string
	PriceCN    string
	LinkUS     string
	LinkHK     string
	LinkCN     string
	Img        string
}

type MotherboardType struct {
	Name       string
	Brand      string
	Socket     string
	Chipset    string
	RamSlot    int
	RamType    string
	RamSupport []string
	RamMax     int
	Pcie16Slot int
	Pcie4Slot  int
	Pcie1Slot  int
	M2Slot     int
	SataSlot   int
	FormFactor string
	Wireless   bool
	PriceUS    string
	PriceHK    string
	PriceCN    string
	LinkUS     string
	LinkHK     string
	LinkCN     string
	Img        string
}

func GetMotherboardSpec(record LinkRecord) MotherboardSpec {

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

	specCollector := collector.Clone()

	motherboardData := getMotherboardSpecData(record.LinkSpec, specCollector)

	if strings.Contains(strings.ToUpper(record.Name), "WIFI") {
		motherboardData.Wireless = true
	}
	motherboardData.Code = record.Name
	motherboardData.Brand = record.Brand
	motherboardData.PriceCN = record.PriceCN
	motherboardData.PriceHK = ""
	motherboardData.LinkHK = ""
	motherboardData.LinkCN = record.LinkCN
	if record.LinkUS != "" {
		motherboardData.LinkUS = record.LinkUS
	}
	return motherboardData
}

func GetMotherboardData(spec MotherboardSpec) MotherboardType {

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

	return MotherboardType{
		Name:       spec.Name,
		Brand:      spec.Brand,
		Socket:     spec.Socket,
		Chipset:    spec.Chipset,
		RamSlot:    spec.RamSlot,
		RamType:    spec.RamType,
		RamSupport: spec.RamSupport,
		RamMax:     spec.RamMax,
		Pcie16Slot: spec.Pcie1Slot,
		Pcie4Slot:  spec.Pcie4Slot,
		Pcie1Slot:  spec.Pcie16Slot,
		M2Slot:     spec.M2Slot,
		SataSlot:   spec.SataSlot,
		FormFactor: spec.FormFactor,
		Wireless:   spec.Wireless,
		LinkUS:     spec.LinkUS,
		LinkHK:     spec.LinkHK,
		LinkCN:     spec.LinkCN,
		PriceCN:    priceCN,
		PriceUS:    priceUS,
		PriceHK:    "",
		Img:        tempImg,
	}
}

func getMotherboardSpecData(link string, collector *colly.Collector) MotherboardSpec {
	specData := MotherboardSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".tns-inner img", "src")
		specData.Name = element.ChildText(".breadcrumb .active")
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

		specData.RamType = element.ChildText(".table-striped .badge-primary")
		var ramSupportList []string
		fmt.Println(specData.PriceUS)

		element.ForEach(".table-striped .ram-values span", func(i int, item *colly.HTMLElement) {
			temp := item.Text
			ramSupportList = append(ramSupportList, temp)
		})

		specData.RamSupport = ramSupportList

		element.ForEach("ul.tail-links a", func(i int, item *colly.HTMLElement) {
			itemStr := item.ChildText("strong")
			if strings.Contains(itemStr, "PCI-Express x16 Slots") {
				specData.Pcie16Slot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "PCI-Express x4 Slots") {
				specData.Pcie4Slot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "PCI-Express x1 Slots") {
				specData.Pcie1Slot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "M.2 Ports") {
				specData.M2Slot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "RAM Slots") {
				specData.RamSlot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "Supported RAM") {
				specData.RamMax = extractNumberFromString(itemStr)
			}
		})

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Socket":
				specData.Socket = item.ChildTexts("td")[1]
			case "Form factor":
				specData.FormFactor = item.ChildTexts("td")[1]
			case "Chipset":
				specData.Chipset = item.ChildTexts("td")[1]
			}
		})
	})

	collector.Visit(link)

	return specData
}
