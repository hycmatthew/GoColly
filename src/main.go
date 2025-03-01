package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"go-colly-lib/src/databaseLogic"
	"go-colly-lib/src/pcData"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
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

	getDataName := pcCase
	isUpdateSpec := true

	if isUpdateSpec {
		if getDataName == gpuScore {
			updateGPUScoreLogic()
		} else {
			updateSpecLogic(getDataName)
		}
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
	fmt.Println("mergeSpecData ", name, " - ", count)
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

func saveRecordToDatabase(part string, record databaseLogic.DBRecord) {
	databaseLogic.InsertRecord(part, record)
}

/*
GPU SPEC
*/
func updateGPUScoreLogic() {
	timeSet := 8000
	timeDuration := time.Duration(timeSet) * time.Millisecond
	ticker := time.NewTicker(timeDuration)

	dataList := readCsvFile("res/gpuspecdata.csv")
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
func updatePriceLogic(name string) {
	timeSet := 5000
	extraTry := 50
	maxRetryTime := 3
	retryTime := 0
	timeDuration := time.Duration(timeSet) * time.Millisecond
	ticker := time.NewTicker(timeDuration)

	specFile, _ := os.Open("tmp/spec/" + name + "Spec.json")
	byteValue, _ := io.ReadAll(specFile)
	count := 0

	dataFile, _ := os.Open("tmp/" + name + "Data.json")
	dataByteValue, _ := io.ReadAll(dataFile)

	switch name {
	case "cpu":
		var specList []pcData.CPUSpec
		var cpuList []pcData.CPUType
		json.Unmarshal([]byte(byteValue), &specList)

		var oldCpuList []pcData.CPUType
		json.Unmarshal([]byte(dataByteValue), &oldCpuList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record, valid := pcData.GetCPUData(spec)
				if valid || retryTime == maxRetryTime {
					result := pcData.CompareCPUDataLogic(record, oldCpuList)
					cpuList = append(cpuList, result)
					retryTime = 0
					count++
				} else {
					retryTime++
				}

				if count == len(specList) {
					saveData(cpuList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + extraTry))
		time.Sleep(time.Second * listLen)
	case "gpu":
		var specList []pcData.GPUScore
		var gpuList []pcData.GPUType
		json.Unmarshal([]byte(byteValue), &specList)

		var oldGpuList []pcData.GPUType
		json.Unmarshal([]byte(dataByteValue), &oldGpuList)

		dataList := readCsvFile("res/" + name + "data.csv")
		var recordList []pcData.GPURecordData
		count++

		for i := 1; i < len(dataList); i++ {
			data := dataList[i]
			record := pcData.GPURecordData{Brand: data[0], Name: data[1], PriceCN: data[2], SpecCN: data[3], LinkCN: data[4], LinkUS: data[5], LinkHK: data[6]}
			recordList = append(recordList, record)
		}

		go func() {
			for {
				<-ticker.C
				data := recordList[count]
				record, valid := pcData.GetGPUData(specList, data)
				if valid || retryTime == maxRetryTime {
					result := pcData.CompareGPUDataLogic(record, oldGpuList)
					gpuList = append(gpuList, result)
					retryTime = 0
					count++
				} else {
					retryTime++
				}

				if count == len(recordList) {
					saveData(gpuList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(recordList) + extraTry))
		time.Sleep(time.Second * listLen)
	case "motherboard":
		var specList []pcData.MotherboardSpec
		var mbList []pcData.MotherboardType
		json.Unmarshal([]byte(byteValue), &specList)

		var oldMotherboardList []pcData.MotherboardType
		json.Unmarshal([]byte(dataByteValue), &oldMotherboardList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record, valid := pcData.GetMotherboardData(spec)
				if valid || retryTime == maxRetryTime {
					result := pcData.CompareMotherboardDataLogic(record, oldMotherboardList)
					mbList = append(mbList, result)
					retryTime = 0
					count++
				} else {
					retryTime++
				}

				if count == len(specList) {
					saveData(mbList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + extraTry))
		time.Sleep(time.Second * listLen)
	case "ram":
		var specList []pcData.RamSpec
		var ramList []pcData.RamType
		json.Unmarshal([]byte(byteValue), &specList)

		var oldRamList []pcData.RamType
		json.Unmarshal([]byte(dataByteValue), &oldRamList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record, valid := pcData.GetRamData(spec)
				if valid || retryTime == maxRetryTime {
					result := pcData.CompareRAMDataLogic(record, oldRamList)
					ramList = append(ramList, result)
					retryTime = 0
					count++
				} else {
					retryTime++
				}

				if count == len(specList) {
					saveData(ramList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + extraTry))
		time.Sleep(time.Second * listLen)
	case "ssd":
		var specList []pcData.SSDSpec
		var ssdList []pcData.SSDType
		json.Unmarshal([]byte(byteValue), &specList)

		var oldSSDList []pcData.SSDType
		json.Unmarshal([]byte(dataByteValue), &oldSSDList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record, valid := pcData.GetSSDData(spec)
				if valid || retryTime == maxRetryTime {
					result := pcData.CompareSSDDataLogic(record, oldSSDList)
					ssdList = append(ssdList, result)
					retryTime = 0
					count++
				} else {
					retryTime++
				}

				if count == len(specList) {
					saveData(ssdList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + extraTry))
		time.Sleep(time.Second * listLen)
	case "case":
		var specList []pcData.CaseSpec
		var caseList []pcData.CaseType
		json.Unmarshal([]byte(byteValue), &specList)

		var oldCaseList []pcData.CaseType
		json.Unmarshal([]byte(dataByteValue), &oldCaseList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record, valid := pcData.GetCaseData(spec)
				if valid || retryTime == maxRetryTime {
					result := pcData.CompareCaseDataLogic(record, oldCaseList)
					caseList = append(caseList, result)
					retryTime = 0
					count++
				} else {
					retryTime++
				}

				if count == len(specList) {
					validationData := pcData.LoadValidationData("case")
					mergedCaseList := pcData.MergeData(caseList, validationData)
					saveData(mergedCaseList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + extraTry))
		time.Sleep(time.Second * listLen)
	case "cooler":
		var specList []pcData.CoolerSpec
		var coolerList []pcData.CoolerType
		json.Unmarshal([]byte(byteValue), &specList)

		var oldCoolerList []pcData.CoolerType
		json.Unmarshal([]byte(dataByteValue), &oldCoolerList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record, valid := pcData.GetCoolerData(spec)
				if valid || retryTime == maxRetryTime {
					result := pcData.CompareCoolerDataLogic(record, oldCoolerList)
					coolerList = append(coolerList, result)
					retryTime = 0
					count++
				} else {
					retryTime++
				}

				if count == len(specList) {
					saveData(coolerList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + extraTry))
		time.Sleep(time.Second * listLen)
	case "power":
		var specList []pcData.PowerSpec
		var powerList []pcData.PowerType
		json.Unmarshal([]byte(byteValue), &specList)

		var oldPowerList []pcData.PowerType
		json.Unmarshal([]byte(dataByteValue), &oldPowerList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record, valid := pcData.GetPowerData(spec)
				if valid || retryTime == maxRetryTime {
					result := pcData.ComparePowerDataLogic(record, oldPowerList)
					powerList = append(powerList, result)
					retryTime = 0
					count++
				} else {
					retryTime++
				}

				if count == len(specList) {
					saveData(powerList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + extraTry))
		time.Sleep(time.Second * listLen)
	default:
		fmt.Println("something wrong!!")
		var specList []pcData.RamSpec
		var ramList []pcData.RamType

		json.Unmarshal([]byte(byteValue), &specList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record, valid := pcData.GetRamData(spec)
				if valid {
					ramList = append(ramList, record)
					count++
				}

				if count == len(specList) {
					saveData(ramList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + extraTry))
		time.Sleep(time.Second * listLen)
	}
}
