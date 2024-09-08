package pcData

type GPUSpecTempStruct struct {
	Name        string
	Generation  string
	MemorySize  string
	MemoryType  string
	MemoryBus   string
	Clock       string
	Power       int
	Length      int
	Slot        string
	Width       int
	ProductSpec []GPUSpecSubData
}

type GPUSpecSubData struct {
	ProductName string
	BoostClock  string
	Length      int
	Slots       string
	TDP         int
}

type GPUSpec struct {
	Name string
	Link string
}

func GetGPUSpecDataList() []GPUSpec {
	list := []GPUSpec{
		{Name: "RTX 4060", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4060.c4107"},
		{Name: "RTX 4060 Ti", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4060-ti-8-gb.c3890"},
		{Name: "RTX 4070", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4070.c3924"},
		{Name: "RTX 4070 SUPER", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4070-super.c4186"},
		{Name: "RTX 4070 Ti", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4070-ti.c3950"},
		{Name: "RTX 4070 Ti SUPER", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4070-ti-super.c4187"},
		{Name: "RTX 4070 Ti SUPER AD102", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4070-ti-super-ad102.c4215"},
		{Name: "RTX 4080", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4080.c3888"},
		{Name: "RTX 4080 SUPER", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4080-super.c4182"},
		{Name: "RTX 4080 Ti", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4080-ti.c3887"},
		{Name: "RTX 4090", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4090.c3889"},
		{Name: "RTX 4090 D", Link: "https://www.techpowerup.com/gpu-specs/geforce-rtx-4090-d.c4189"},
	}

	return list
}
