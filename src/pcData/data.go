package pcData

type GPUSpecTempStruct struct {
	Name        string
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

/*
func getdGPUScore(generation string) int {
	scoreRes := 0

	scoreMap := []gpuScore{
		{model: "RTX 4090", score: 36499},
		{model: "RTX 4090 D", score: 34332},
		{model: "RX 7900 XTX", score: 30621},
		{model: "RTX 4080 SUPER", score: 28371},
		{model: "RTX 4080", score: 28272},
		{model: "RX 7900 XT", score: 26911},
		{model: "RTX 4070 Ti SUPER", score: 24253},
		{model: "RTX 4070 Ti", score: 22832},
		{model: "RX 7900 GRE", score: 22344},
		{model: "RTX 4070 SUPER", score: 20968},
		{model: "RX 7800 XT", score: 20031},
		{model: "RTX 4070", score: 17856},
		{model: "RX 7700 XT", score: 16991},
		{model: "RTX 4060 Ti", score: 13509},
		{model: "RX 7600 XT", score: 11296},
		{model: "RX 7600", score: 11014},
		{model: "RTX 4060", score: 10620},
	}

	for _, v := range scoreMap {
		if v.model == generation {
			scoreRes = v.score
		}
	}
	return scoreRes
}
*/
