package pcData

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

type MotherboardRecord struct {
	Name   string
	Spec   string
	LinkCN string
	LinkHK string
	LinkUS string
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
	PriceUS    float64
	PriceHK    float64
	PriceCN    float64
	Img        string
}

func GetMotherboardData(name string, spec string, cnLink string, enLink string, hkLink string) MotherboardType {

	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			// "https://nanoreview.net",
			"nanoreview.net",
			"www.newegg.com",
			"newegg.com",
			// "https://cu.manmanbuy.com",
			"cu.manmanbuy.com",
			"www.price.com.hk",
			"price.com.hk",
			"detail.zol.com.cn",
			"zol.com.cn",
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
	usCollector := collector.Clone()
	cnCollector := collector.Clone()
	// hkCollector := collector.Clone()

	motherboardData := getMotherboardSpec(spec, specCollector)
	motherboardData.PriceUS = getMotherboardUSPrice(enLink, usCollector)
	motherboardData.PriceCN = getMotherboardCNPrice(cnLink, cnCollector)
	// motherboardData.PriceHK = getMotherboardHKPrice(hkLink, hkCollector)

	if strings.Contains(strings.ToUpper(name), "WIFI") {
		motherboardData.Wireless = true
	}

	return motherboardData
}

func getMotherboardSpec(link string, collector *colly.Collector) MotherboardType {
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
	price := 0.0
	imgLink := ""

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".tns-inner img", "src")

		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current")), 64); err == nil {
			price = s
		}

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
	/*
		fmt.Println(MotherboardType{
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
			Wireless:   wireless,
			PriceUS:    price,
			PriceHK:    0,
			PriceCN:    0,
			Img:        imgLink,
		})
	*/
	return MotherboardType{
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
		PriceHK:    0,
		PriceCN:    0,
		Img:        imgLink,
	}
}

func getMotherboardUSPrice(link string, collector *colly.Collector) float64 {
	price := 0.0

	collectorErrorHandle(collector, link)
	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		if s, err := strconv.ParseFloat(extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current")), 64); err == nil {
			price = s
		}
	})

	collector.Visit(link)
	return price
}

func getMotherboardHKPrice(link string, collector *colly.Collector) float64 {
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

func getMotherboardCNPrice(link string, collector *colly.Collector) float64 {
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
