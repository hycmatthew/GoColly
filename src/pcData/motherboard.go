package pcData

import (
	"fmt"
	"net/http"
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
	RamSupport string
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
	RamSupport string
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
	if priceCN != "" {
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
	name := ""
	brand := ""
	socket := ""
	chipset := ""
	ramSlot := 0
	ramType := ""
	ramSupport := ""
	ramMax := 0
	pcie16Slot := 0
	pcie4Slot := 0
	pcie1Slot := 0
	m2Slot := 0
	sataSlot := 0
	formFactor := ""
	price := ""
	imgLink := ""

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".tns-inner img", "src")

		price = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current"))
		ramType = element.ChildText(".table-striped .badge-primary")
		var ramSupportList []string

		element.ForEach(".table-striped .ram-values span", func(i int, item *colly.HTMLElement) {
			temp := item.Text
			ramSupportList = append(ramSupportList, temp)
		})

		for _, str := range ramSupportList {
			ramSupport += (str + " ")
		}
		fmt.Println(ramSupport)

		element.ForEach("ul.tail-links a", func(i int, item *colly.HTMLElement) {
			itemStr := item.ChildText("strong")
			if strings.Contains(itemStr, "Socket") {
				socket = itemStr
			}
			if strings.Contains(itemStr, "Form factor") {
				formFactor = itemStr
			}
			if strings.Contains(itemStr, "Chipset") {
				chipset = itemStr
			}
			if strings.Contains(itemStr, "PCI-Express x16 Slots") {
				pcie16Slot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "PCI-Express x4 Slots") {
				pcie4Slot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "PCI-Express x1 Slots") {
				pcie1Slot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "M.2 Ports") {
				m2Slot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "RAM Slots") {
				ramSlot = extractNumberFromString(itemStr)
			}
			if strings.Contains(itemStr, "Supported RAM") {
				ramMax = extractNumberFromString(itemStr)
			}
		})
	})

	collector.Visit(link)

	return MotherboardSpec{
		Name:       name,
		Brand:      brand,
		Socket:     socket,
		Chipset:    chipset,
		RamSlot:    ramSlot,
		RamType:    ramType,
		RamSupport: ramSupport,
		RamMax:     ramMax,
		Pcie16Slot: pcie16Slot,
		Pcie4Slot:  pcie4Slot,
		Pcie1Slot:  pcie1Slot,
		M2Slot:     m2Slot,
		SataSlot:   sataSlot,
		FormFactor: formFactor,
		Wireless:   false,
		PriceUS:    price,
		PriceHK:    "",
		PriceCN:    "",
		Img:        imgLink,
	}
}
