package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"go-colly-lib/src/pcData"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

func main() {
	const (
		cpu         = "cpu"
		gpu         = "gpu"
		motherboard = "motherboard"
		ram         = "ram"
		ssd         = "ssd"
		power       = "power"
		cooler      = "cooler"
		pcCase      = "case"
		gpuScore    = "gpuScore"
	)

	getDataName := cpu
	isUpdateSpec := false

	if getDataName == gpuScore {
		if isUpdateSpec {
			updateGPUScoreLogic()
		} else {
			UpdateBenchmarks()
		}
	} else if isUpdateSpec {
		updateSpecLogic(getDataName)
	} else {
		updatePriceLogic(getDataName)
	}
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

func saveData(result any, name string) {
	fmt.Println("save Data started!")
	jsonData, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write JSON data to file
	err = os.WriteFile("tmp/"+name+"Data.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

// Merge Spec Logic
func readSpecData(name string) ([]map[string]any, error) {
	dirPath := "tmp/spec/"
	filename := name + "Spec.json"
	filePath := filepath.Join(dirPath, filename)
	if _, err := os.ReadFile(filePath); errors.Is(err, os.ErrNotExist) {
		return []map[string]any{}, nil // 文件不存在，返回空数组
	}
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var data []map[string]any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func isSlice(v any) bool {
	value := reflect.ValueOf(v)
	return value.Kind() == reflect.Slice
}

func structToMap(data any) map[string]any {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Struct {
		panic("Input must be a struct")
	}
	m := make(map[string]any)
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if field.PkgPath != "" { // 跳过非导出字段
			continue
		}
		m[field.Name] = val.Field(i).Interface()
	}
	return m
}

func resultToMaps(result any) []map[string]any {
	if !isSlice(result) {
		// 单个结构，包装为单元素数组
		singleMap := structToMap(result)
		return []map[string]any{singleMap}
	}
	sliceVal := reflect.ValueOf(result)
	maps := make([]map[string]any, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i).Interface()
		maps[i] = structToMap(elem)
	}
	return maps
}

func updateData(existing []map[string]any, newMaps []map[string]any) []map[string]any {
	codeIndexMap := make(map[string]int)
	for i, m := range existing {
		code, ok := m["Code"].(string)
		if !ok {
			panic("Invalid existing data, 'Code' is not a string")
		}
		codeIndexMap[code] = i
	}
	for _, newMap := range newMaps {
		code, ok := newMap["Code"].(string)
		if !ok {
			panic("Invalid new data, 'Code' is not a string")
		}
		if idx, ok := codeIndexMap[code]; ok {
			existing[idx] = newMap
		} else {
			existing = append(existing, newMap)
		}
	}
	return existing
}

func mergeSpecData(result any, name string, count int) {
	fmt.Println("mergeSpecData", name, " - ", count)
	existing, err := readSpecData(name)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	newMaps := resultToMaps(result)
	updated := updateData(existing, newMaps)
	err = saveSpecData(updated, name)
	if err != nil {
		fmt.Println("Error writing file:", err)
	}
}

func saveSpecData(result any, name string) error {
	fmt.Println("save spec started!")
	jsonData, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Write JSON data to file
	err = os.WriteFile("tmp/spec/"+name+"Spec.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return nil
}

/*
GPU SPEC
*/
func updateGPUScoreLogic() {
	timeSet := 8000
	timeDuration := time.Duration(timeSet) * time.Millisecond
	ticker := time.NewTicker(timeDuration)

	dataList := readCsvFile("res/gpuscoredata.csv")
	var recordList []pcData.GPUScoreData

	for i := 1; i < len(dataList); i++ {
		data := dataList[i]
		record := pcData.GPUScoreData{Name: data[0], Benchmark: data[1], DataLink: data[2]}
		recordList = append(recordList, record)
	}

	count := 0
	var specList []pcData.GPUScore
	go func() {
		for {
			<-ticker.C
			gpuRecord := pcData.GetGPUScoreSpec(recordList[count])
			specList = append(specList, gpuRecord)
			count++
			if count == len(recordList) {
				saveSpecData(specList, "gpuScore")
				ticker.Stop()
				runtime.Goexit()
			}
		}
	}()

	listLen := time.Duration(timeSet * (len(recordList) + 3))
	time.Sleep(time.Second * listLen)
}

func UpdateBenchmarks() {
	dataList := loadJSON[[]pcData.GPUType]("tmp/gpuData.json")
	var recordList []pcData.GPUType

	for i := 0; i < len(dataList); i++ {
		data := dataList[i]
		record := pcData.UpdateGPUBenchmarks(data)
		recordList = append(recordList, record)

	}
	saveData(recordList, "gpu")
}

// Update parts spec
type specHandler[T any] struct {
	getSpecFunc func(pcData.LinkRecord) T
	specList    []T
}

func updateSpecLogic(name string) {
	timeSet := 5000
	extraTry := 50
	timeDuration := time.Duration(timeSet) * time.Millisecond
	ticker := time.NewTicker(timeDuration)

	dataList := readCsvFile("res/" + name + "data.csv")
	var recordList []pcData.LinkRecord

	for i := 1; i < len(dataList); i++ {
		data := dataList[i]
		record := pcData.LinkRecord{
			Brand:    data[0],
			Name:     data[1],
			PriceCN:  data[2],
			LinkSpec: data[3],
			LinkCN:   data[4],
			LinkUS:   data[5],
			LinkHK:   data[6],
		}
		recordList = append(recordList, record)
	}

	handlers := map[string]interface{}{
		"cpu": specHandler[pcData.CPUSpec]{
			getSpecFunc: pcData.GetCPUSpec,
			specList:    []pcData.CPUSpec{},
		},
		"gpu": specHandler[pcData.GPUSpec]{
			getSpecFunc: pcData.GetGPUSpec,
			specList:    []pcData.GPUSpec{},
		},
		"motherboard": specHandler[pcData.MotherboardSpec]{
			getSpecFunc: pcData.GetMotherboardSpec,
			specList:    []pcData.MotherboardSpec{},
		},
		"ram": specHandler[pcData.RamSpec]{
			getSpecFunc: pcData.GetRamSpec,
			specList:    []pcData.RamSpec{},
		},
		"ssd": specHandler[pcData.SSDSpec]{
			getSpecFunc: pcData.GetSSDSpec,
			specList:    []pcData.SSDSpec{},
		},
		"power": specHandler[pcData.PowerSpec]{
			getSpecFunc: pcData.GetPowerSpec,
			specList:    []pcData.PowerSpec{},
		},
		"case": specHandler[pcData.CaseSpec]{
			getSpecFunc: pcData.GetCaseSpec,
			specList:    []pcData.CaseSpec{},
		},
		"cooler": specHandler[pcData.CoolerSpec]{
			getSpecFunc: pcData.GetCoolerSpec,
			specList:    []pcData.CoolerSpec{},
		},
	}

	if handler, ok := handlers[name]; ok {
		go func(h interface{}) {
			switch h := h.(type) {
			case specHandler[pcData.CPUSpec]:
				processSpecs(h, recordList, ticker, name)
			case specHandler[pcData.GPUSpec]:
				processSpecs(h, recordList, ticker, name)
			case specHandler[pcData.MotherboardSpec]:
				processSpecs(h, recordList, ticker, name)
			case specHandler[pcData.RamSpec]:
				processSpecs(h, recordList, ticker, name)
			case specHandler[pcData.SSDSpec]:
				processSpecs(h, recordList, ticker, name)
			case specHandler[pcData.PowerSpec]:
				processSpecs(h, recordList, ticker, name)
			case specHandler[pcData.CaseSpec]:
				processSpecs(h, recordList, ticker, name)
			case specHandler[pcData.CoolerSpec]:
				processSpecs(h, recordList, ticker, name)
			}
		}(handler)
	}

	listLen := time.Duration(timeSet * (len(recordList) + extraTry))
	time.Sleep(time.Second * listLen)
}

func processSpecs[T any](handler specHandler[T], records []pcData.LinkRecord, ticker *time.Ticker, name string) {
	count := 0
	for range ticker.C {
		if count >= len(records) {
			mergeSpecData(handler.specList, name, count)
			ticker.Stop()
			return
		}
		if count%50 == 0 && count != 0 {
			mergeSpecData(handler.specList, name, count)
		}
		spec := handler.getSpecFunc(records[count])
		if !isZero(spec) {
			handler.specList = append(handler.specList, spec)
			count++
		}
	}
}

func isZero[T any](v T) bool {
	return reflect.ValueOf(&v).Elem().IsZero()
}

// update Price Logic
type Processor[T any, D any] struct {
	GetData       func(T) (D, bool)
	ProcessResult func([]D, map[string][]string) []D
}

func updatePriceLogic(name string) {
	// 配置所有硬件类型的处理参数
	processors := map[string]interface{}{
		"cpu": Processor[pcData.CPUSpec, pcData.CPUType]{
			GetData: pcData.GetCPUData,
			ProcessResult: func(data []pcData.CPUType, vd map[string][]string) []pcData.CPUType {
				return pcData.MergeData(data, vd)
			},
		},
		"gpu": Processor[pcData.GPUSpec, pcData.GPUType]{
			GetData: pcData.GetGPUData,
			ProcessResult: func(data []pcData.GPUType, vd map[string][]string) []pcData.GPUType {
				return pcData.MergeData(data, vd)
			},
		},
		"motherboard": Processor[pcData.MotherboardSpec, pcData.MotherboardType]{
			GetData: pcData.GetMotherboardData,
			ProcessResult: func(data []pcData.MotherboardType, vd map[string][]string) []pcData.MotherboardType {
				return pcData.MergeData(data, vd)
			},
		},
		"ram": Processor[pcData.RamSpec, pcData.RamType]{
			GetData: pcData.GetRamData,
			ProcessResult: func(data []pcData.RamType, vd map[string][]string) []pcData.RamType {
				return pcData.MergeData(data, vd)
			},
		},
		"ssd": Processor[pcData.SSDSpec, pcData.SSDType]{
			GetData: pcData.GetSSDData,
			ProcessResult: func(data []pcData.SSDType, vd map[string][]string) []pcData.SSDType {
				return pcData.MergeData(data, vd)
			},
		},
		"power": Processor[pcData.PowerSpec, pcData.PowerType]{
			GetData: pcData.GetPowerData,
			ProcessResult: func(data []pcData.PowerType, vd map[string][]string) []pcData.PowerType {
				return pcData.MergeData(data, vd)
			},
		},
		"case": Processor[pcData.CaseSpec, pcData.CaseType]{
			GetData: pcData.GetCaseData,
			ProcessResult: func(data []pcData.CaseType, vd map[string][]string) []pcData.CaseType {
				return pcData.MergeData(data, vd)
			},
		},
		"cooler": Processor[pcData.CoolerSpec, pcData.CoolerType]{
			GetData: pcData.GetCoolerData,
			ProcessResult: func(data []pcData.CoolerType, vd map[string][]string) []pcData.CoolerType {
				return pcData.MergeData(data, vd)
			},
		},
	}

	// 获取具体处理器配置
	processor, ok := processors[name]
	if !ok {
		fmt.Printf("Unsupported hardware type: %s\n", name)
		return
	}

	// 通用处理流程
	switch p := processor.(type) {
	case Processor[pcData.CPUSpec, pcData.CPUType]:
		processGeneric[pcData.CPUSpec, pcData.CPUType](name, p)
	case Processor[pcData.GPUSpec, pcData.GPUType]:
		processGeneric[pcData.GPUSpec, pcData.GPUType](name, p)
	case Processor[pcData.MotherboardSpec, pcData.MotherboardType]:
		processGeneric[pcData.MotherboardSpec, pcData.MotherboardType](name, p)
	case Processor[pcData.RamSpec, pcData.RamType]:
		processGeneric[pcData.RamSpec, pcData.RamType](name, p)
	case Processor[pcData.SSDSpec, pcData.SSDType]:
		processGeneric[pcData.SSDSpec, pcData.SSDType](name, p)
	case Processor[pcData.PowerSpec, pcData.PowerType]:
		processGeneric[pcData.PowerSpec, pcData.PowerType](name, p)
	case Processor[pcData.CaseSpec, pcData.CaseType]:
		processGeneric[pcData.CaseSpec, pcData.CaseType](name, p)
	case Processor[pcData.CoolerSpec, pcData.CoolerType]:
		processGeneric[pcData.CoolerSpec, pcData.CoolerType](name, p)
	}
}

// 泛型处理核心逻辑
func processGeneric[T any, D any](
	name string,
	processor Processor[T, D],
) {
	const (
		timeSet      = 5000
		extraTry     = 50
		maxRetryTime = 3
	)
	// 加载数据
	specList := loadJSON[[]T]("tmp/spec/" + name + "Spec.json")
	filterList := FilterSpecList(specList, name)
	oldData := loadJSON[[]D]("tmp/" + name + "Data.json")

	ticker := time.NewTicker(time.Duration(timeSet) * time.Millisecond)
	defer ticker.Stop()

	var (
		results    []D
		count      int
		retryCount int
	)

	go func() {
		for range ticker.C {
			if count >= len(filterList) {
				validationData := pcData.LoadValidationData(name)
				finalData := processor.ProcessResult(results, validationData)
				saveData(finalData, name)
				runtime.Goexit()
			}

			spec := filterList[count]
			record, valid := processor.GetData(spec)

			if valid || retryCount >= maxRetryTime {
				merged := pcData.ComparePreviousDataLogic(record, oldData)
				results = append(results, merged)
				retryCount = 0
				count++
			} else {
				retryCount++
			}
		}
	}()

	sleepDuration := time.Duration(timeSet*(len(filterList)+extraTry)) * time.Millisecond
	time.Sleep(sleepDuration)
}

// 安全加载JSON的泛型函数
func loadJSON[T any](path string) T {
	file, err := os.Open(path)
	if err != nil {
		return *new(T)
	}
	defer file.Close()

	var data T
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return *new(T)
	}
	return data
}

// 泛型过滤函数
func FilterSpecList[T any](specs []T, hardwareType string) []T {
	var filtered []T
	filterRules := map[string][]string{
		"gpu": {"3050", "3060", "3070", "3080", "3090"},
		// 可扩展其他硬件类型的过滤规则
	}

	rules, exists := filterRules[strings.ToLower(hardwareType)]
	if !exists {
		return specs // 未知类型不进行过滤
	}

	for _, spec := range specs {
		v := reflect.ValueOf(spec)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		field := v.FieldByName("Code")
		if !field.IsValid() || field.Kind() != reflect.String {
			continue
		}

		name := strings.ToLower(field.String())
		if !pcData.ContainsAny(name, rules) {
			filtered = append(filtered, spec)
		}
	}
	return filtered
}
