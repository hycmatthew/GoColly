package pcData

import (
	"strings"
)

type GPUSpecTempStruct struct {
	Series      string
	Generation  string
	MemorySize  string
	MemoryType  string
	MemoryBus   string
	Clock       int
	Score       int
	Power       int
	Length      int
	Slot        string
	Width       int
	ProductSpec []GPUSpecSubData
}

type GPUSpecSubData struct {
	ProductName string
	BoostClock  int
	Length      int
	Slots       string
	TDP         int
}

type GPUSpec struct {
	Name  string
	Link  string
	Score int
}

func GetGPUSpecDataList() []GPUSpec {
	list := []GPUSpec{
		{Name: "RTX 4060", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4060.c4107", Score: 10620},
		{Name: "RTX 4060 Ti", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4060-ti-8-gb.c3890", Score: 13509},
		{Name: "RTX 4070", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4070.c3924", Score: 17856},
		{Name: "RTX 4070 SUPER", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4070-super.c4186", Score: 20968},
		{Name: "RTX 4070 Ti", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4070-ti.c3950", Score: 22832},
		{Name: "RTX 4070 Ti SUPER", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4070-ti-super.c4187", Score: 24253},
		{Name: "RTX 4070 Ti SUPER AD102", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4070-ti-super-ad102.c4215", Score: 24253},
		{Name: "RTX 4080", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4080.c3888", Score: 28272},
		{Name: "RTX 4080 SUPER", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4080-super.c4182", Score: 28371},
		{Name: "RTX 4090", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4090.c3889", Score: 36499},
		{Name: "RTX 4090 D", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4090-d.c4189", Score: 34332},
		{Name: "RX 7600", Link: "https://www.techpowerup.com/gpu-specs/radeon-rx-7600.c4153", Score: 0},
		{Name: "RX 7600 XT", Link: "https://www.techpowerup.com/gpu-specs/radeon-rx-7600-xt.c4190", Score: 0},
		{Name: "RX 7700 XT", Link: "https://www.techpowerup.com/gpu-specs/radeon-rx-7700-xt.c3911", Score: 0},
		{Name: "RX 7800 XT", Link: "https://www.techpowerup.com/gpu-specs/radeon-rx-7800-xt.c3839", Score: 0},
		{Name: "RX 7900 GRE", Link: "https://www.techpowerup.com/gpu-specs/radeon-rx-7900-gre.c4166", Score: 0},
		{Name: "RX 7900 XT", Link: "https://www.techpowerup.com/gpu-specs/radeon-rx-7900-xt.c3912", Score: 0},
		{Name: "RX 7900 XTX", Link: "https://www.techpowerup.com/gpu-specs/radeon-rx-7900-xtx.c3941", Score: 0},
	}

	return list
}

func filterByBrand(brand string, in []GPUSpecSubData) []GPUSpecSubData {
	var out []GPUSpecSubData
	for _, each := range in {
		brandStr := strings.Split(each.ProductName, " ")[0]
		if strings.ToLower(brandStr) == brand {
			out = append(out, each)
		}
	}
	return out
}

func getBrandSeries(brand string) [][]string {
	asusSeries := [][]string{
		{"DUAL", "V2"},
		{"DUAL", "WHITE"},
		{"MEGALODON"},
		{"PRIME"},
		{"STRIX"},
		{"TUF"},
	}
	colorfulSeries := [][]string{
		{"iGame", "Advanced"},
		{"iGame", "Ultra", "DUO"},
		{"iGame", "Ultra"},
		{"Tomahawk", "Deluxe"},
		{"Tomahawk", "DUO"},
	}
	galaxySeries := [][]string{
		{"Click"},
		{"BOOMSTAR"},
		{"EX"},
		{"EX", "White"},
		{"METALTOP"},
	}

	gigabyteSeries := [][]string{
		{"AORUS", "ELITE"},
		{"AERO"},
		{"EAGLE"},
		{"GAMING"},
		{"WINDFORCE"},
	}

	msiSeries := [][]string{
		{"GAMING"},
		{"GAMING", "TRIO"},
		{"GAMING", "X"},
		{"GAMING", "X", "TRIO"},
		{"VENTUS", "2X"},
		{"VENTUS", "3X"},
	}

	switch brand {
	case "asus":
		return asusSeries
	case "colorful":
		return colorfulSeries
	case "galaxy":
		return galaxySeries
	case "gigabyte":
		return gigabyteSeries
	default:
		return msiSeries
	}

}

func searchSubDataByName(name string, brand string, subDataList []GPUSpecSubData) GPUSpecSubData {
	brandStr := strings.ToLower(brand)
	seriesList := getBrandSeries(brandStr)
	for i := range seriesList {
		for j := range seriesList[i] {
			seriesList[i][j] = strings.ToUpper(seriesList[i][j])
		}
	}

	updatedName := strings.ToUpper(strings.Replace(name, "-", " ", -1))
	nameList := strings.Split(updatedName, " ")
	var matchedseries []string
	isOC := false
	for _, item := range nameList {
		if brandStr == "asus" {
			first := item[0:]
			last := item[len(item)-1:]
			if first == "O" && last == "G" {
				isOC = true
			}
		}
		if item == "OC" {
			isOC = true
		}
	}
	if isOC {
		nameList = append(nameList, "OC")
	}
	for i := range seriesList {
		if isSubset(seriesList[i], nameList) {
			matchedseries = seriesList[i]
		}
	}
	var out GPUSpecSubData
	tempSubdDataList := filterByBrand(brandStr, subDataList)
	for i := range tempSubdDataList {
		upperName := strings.ToUpper(subDataList[i].ProductName)
		subDataNameList := strings.Split(upperName, " ")
		subOC := strings.Contains(upperName, " OC")
		if isSubset(matchedseries, subDataNameList) && isOC == subOC {
			out = subDataList[i]
		}
	}
	return out
}

func isSubset(arr1, arr2 []string) bool {
	set := make(map[string]bool)

	for _, str := range arr2 {
		set[str] = true
	}

	for _, str := range arr1 {
		if !set[str] {
			return false
		}
	}
	return true
}
