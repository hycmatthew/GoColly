package pcData

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
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

type PriceType struct {
	Region    string
	Platform  string
	Price     string
	PriceLink string
}

type CPUSpec struct {
	Code                    string
	Name                    string
	NameCN                  string
	Brand                   string
	Socket                  string
	Cores                   int
	Threads                 int
	GPU                     string
	SingleCoreScore         int
	MultiCoreScore          int
	IntegratedGraphicsScore int
	Power                   int
	Prices                  []PriceType
	Img                     string
}

type CPUType struct {
	Id                      string
	Name                    string
	NameCN                  string
	Brand                   string
	Socket                  string
	Cores                   int
	Threads                 int
	GPU                     string
	SingleCoreScore         int
	MultiCoreScore          int
	IntegratedGraphicsScore int
	Power                   int
	Prices                  []PriceType
	Img                     string
}

func GetCPUSpec(record LinkRecord) CPUSpec {
	cpuData := manualCPUSpecHandle(record.Name)
	if cpuData.Name == "" {
		cpuData = getCPUSpecData(record.LinkSpec, CreateCollector())
	}
	cpuData.Code = record.Name
	cpuData.Name = RemoveBrandsFromName(cpuData.Brand, cpuData.Name)

	// 添加各區域價格連結
	cpuData.Prices = handleSpecPricesLogic(cpuData.Prices, record)
	return cpuData
}

func GetCPUData(spec CPUSpec) (CPUType, bool) {
	isValid := true
	updatedPrices := make([]PriceType, len(spec.Prices))
	copy(updatedPrices, spec.Prices)

	// 建立價格處理map
	priceHandlers := map[string]func(string, *colly.Collector) (string, string){
		"CN": func(link string, c *colly.Collector) (string, string) {
			if strings.Contains(link, "pconline") {
				name, price := getCNNameAndPriceFromPcOnline(link, c)
				return price, name // 返回價格
			}
			return spec.Name, ""
		},
		"US": func(link string, c *colly.Collector) (string, string) {
			if strings.Contains(link, "newegg") {
				price, img := getUSPriceAndImgFromNewEgg(link, c)
				return price, img
			}
			return "", ""
		},
	}

	// 處理每個價格來源
	for i, price := range updatedPrices {
		if price.Price == "" {
			handler, exists := priceHandlers[price.Region]
			if exists {
				collectedPrice, collectedData := handler(price.PriceLink, CreateCollector())
				// 更新價格
				updatedPrices[i].Price = collectedPrice
				// 處理特殊情況
				switch price.Region {
				case "CN":
					if collectedData != "" { // 這裡 collectedData 是中文名稱
						spec.NameCN = RemoveBrandsFromName(spec.Brand, collectedData)
					}
				case "US":
					if collectedData != "" { // 這裡 collectedData 是圖片URL
						spec.Img = collectedData
					}
				}
			}

			if updatedPrices[i].Price == "" {
				isValid = false
			}
		}
	}

	iGPU, iScore := integratedGraphicsScoreHandle(spec.Code)

	return CPUType{
		Id:                      SetProductId(spec.Brand, spec.Code),
		Name:                    spec.Name,
		NameCN:                  spec.NameCN,
		Brand:                   strings.ToLower(spec.Brand),
		Cores:                   spec.Cores,
		Threads:                 spec.Threads,
		Socket:                  spec.Socket,
		GPU:                     iGPU,
		SingleCoreScore:         spec.SingleCoreScore,
		MultiCoreScore:          spec.MultiCoreScore,
		IntegratedGraphicsScore: iScore,
		Power:                   spec.Power,
		Prices:                  updatedPrices,
		Img:                     spec.Img,
	}, isValid
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
		Brand:           strings.ToLower(brand),
		Cores:           cores,
		Threads:         thread,
		Socket:          strings.Replace(socket, "-", "", -1),
		GPU:             gpu,
		SingleCoreScore: singleCoreScore,
		MultiCoreScore:  muitiCoreScore,
		Power:           tdp,
	}
}

func manualCPUSpecHandle(code string) CPUSpec {
	if code == "Core i5-14490F" {
		return CPUSpec{
			Name:            "Intel Core i5 14490F",
			Brand:           "Intel",
			Cores:           10,
			Threads:         16,
			Socket:          "LGA1700",
			GPU:             "No",
			SingleCoreScore: 1899,
			MultiCoreScore:  17396,
			Power:           148,
		}
	}
	return CPUSpec{}
}

// GPUInfo 存储集成显卡信息
type GPUProfile struct {
	Name  string
	Score int
}

// CPU数据库映射表
var (
	// Intel 显卡配置
	UHD730      = GPUProfile{"UHD Graphics 730", 589}
	UHD770      = GPUProfile{"UHD Graphics 770", 816}
	ArcXeLPG    = GPUProfile{"Arc Xe-LPG", 2000}           // 基础版
	ArcXeLPG_LP = GPUProfile{"Arc Xe-LPG (1800MHZ)", 1500} // 低功耗版
	ArcXeLPG_H  = GPUProfile{"Arc Xe-LPG", 2500}           // 高性能版
	ArcXeLPG_T  = GPUProfile{"Arc Xe-LPG", 3000}           // 旗舰版

	// AMD 显卡配置
	Vega_7     = GPUProfile{"Vega 7", 1200}
	Vega_8     = GPUProfile{"Vega 8", 1400}
	RDNA2_2CU  = GPUProfile{"Radeon 2CU (RDNA2)", 1000}
	RDNA3_2CU  = GPUProfile{"Radeon 2CU (RDNA3)", 1200}
	RDNA3_12CU = GPUProfile{"Radeon 12CU (RDNA3)", 3000}

	// 特殊值
	NoGraphics = GPUProfile{"No", 0}
)

// CPU数据库映射表
var cpuDatabase = map[string]GPUProfile{
	// Intel 12/13/14代
	"Core i3-12100":   UHD730,
	"Core i3-12100F":  NoGraphics,
	"Core i3-13100":   UHD730,
	"Core i3-13100F":  NoGraphics,
	"Core i3-14100":   UHD730,
	"Core i3-14100F":  NoGraphics,
	"Core i5-12400":   UHD730,
	"Core i5-12400F":  NoGraphics,
	"Core i5-12500":   UHD770,
	"Core i5-12600K":  UHD770,
	"Core i5-12600KF": NoGraphics,
	"Core i5-13400":   UHD730,
	"Core i5-13400F":  NoGraphics,
	"Core i5-13500":   UHD770,
	"Core i5-13600K":  UHD770,
	"Core i5-13600KF": NoGraphics,
	"Core i5-14400":   UHD730,
	"Core i5-14400F":  NoGraphics,
	"Core i5-14490F":  NoGraphics,
	"Core i5-14500":   UHD770,
	"Core i5-14600K":  UHD770,
	"Core i5-14600KF": NoGraphics,

	// Intel 高端系列
	"Core i7-12700":   UHD770,
	"Core i7-12700F":  NoGraphics,
	"Core i7-12700K":  UHD770,
	"Core i7-12700KF": NoGraphics,
	"Core i7-13700":   UHD770,
	"Core i7-13700F":  NoGraphics,
	"Core i7-13700K":  UHD770,
	"Core i7-13700KF": NoGraphics,
	"Core i7-13900F":  NoGraphics,
	"Core i7-14700":   UHD770,
	"Core i7-14700F":  NoGraphics,
	"Core i7-14700K":  UHD770,
	"Core i7-14700KF": NoGraphics,
	"Core i9-13900":   UHD770,
	"Core i9-13900K":  UHD770,
	"Core i9-13900KF": NoGraphics,
	"Core i9-14900":   UHD770,
	"Core i9-14900K":  UHD770,
	"Core i9-14900KF": NoGraphics,

	// Intel Core Ultra
	"Core Ultra 5 225":   ArcXeLPG_LP,
	"Core Ultra 5 245K":  ArcXeLPG,
	"Core Ultra 5 245KF": NoGraphics,
	"Core Ultra 7 265F":  NoGraphics,
	"Core Ultra 7 265K":  ArcXeLPG_H,
	"Core Ultra 7 265KF": NoGraphics,
	"Core Ultra 9 285K":  ArcXeLPG_T,

	// AMD Ryzen
	"Ryzen 5 5600":    NoGraphics,
	"Ryzen 5 5600G":   Vega_7,
	"Ryzen 5 5600X":   NoGraphics,
	"Ryzen 5 5700G":   Vega_8,
	"Ryzen 5 7600X":   RDNA2_2CU,
	"Ryzen 5 9600X":   RDNA3_2CU,
	"Ryzen 7 5700X":   NoGraphics,
	"Ryzen 7 5700X3D": NoGraphics,
	"Ryzen 7 5800X":   NoGraphics,
	"Ryzen 7 5800X3D": NoGraphics,
	"Ryzen 7 7700X":   RDNA2_2CU,
	"Ryzen 7 7800X3D": RDNA2_2CU,
	"Ryzen 7 8700G":   RDNA3_12CU,
	"Ryzen 7 9700X":   RDNA3_2CU,
	"Ryzen 7 9800X3D": RDNA3_2CU,
	"Ryzen 9 5900X":   NoGraphics,
	"Ryzen 9 5950X":   NoGraphics,
	"Ryzen 9 7900":    RDNA2_2CU,
	"Ryzen 9 7900X":   NoGraphics,
	"Ryzen 9 7900X3D": NoGraphics,
	"Ryzen 9 7950X":   NoGraphics,
	"Ryzen 9 7950X3D": NoGraphics,
	"Ryzen 9 9900X":   RDNA3_2CU,
	"Ryzen 9 9900X3D": RDNA3_2CU,
	"Ryzen 9 9950X":   RDNA3_2CU,
	"Ryzen 9 9950X3D": RDNA3_2CU,
}

// GetIntegratedGraphics 获取集成显卡信息
func integratedGraphicsScoreHandle(cpuName string) (gpu string, score int) {
	info, exists := cpuDatabase[cpuName]
	if !exists {
		fmt.Printf("error: CPU型号 '%s' 不存在数据库\n", cpuName)
		return "error", 0
	}
	return info.Name, info.Score
}
