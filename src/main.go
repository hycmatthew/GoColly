package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"go-colly-lib/src/pcData"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

func main() {
	const (
		cpu         = "cpu"
		gpu         = "gpu"
		motherboard = "mb"
		ram         = "ram"
	)

	getDataName := ram
	isUpdateSpec := true

	if isUpdateSpec {
		if getDataName == "gpu" {
			updateGPUSpecLogic()
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
	fmt.Println("saveData started")
	jsonData, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write JSON data to file
	err = os.WriteFile("tmp/"+name+".json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func saveSpecData(result any, name string) {
	fmt.Println("save spec started!")
	jsonData, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write JSON data to file
	err = os.WriteFile("tmp/spec/"+name+"Spec.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

/*
GPU SPEC
*/
func updateGPUSpecLogic() {
	timeSet := 3000
	timeDuration := time.Duration(timeSet) * time.Millisecond
	ticker := time.NewTicker(timeDuration)

	dataList := readCsvFile("res/gpuspecdata.csv")
	var recordList []pcData.GPUScoreData

	for i := 1; i < len(dataList); i++ {
		data := dataList[i]
		record := pcData.GPUScoreData{Name: data[0], ScoreLink: data[1], DataLink: data[2]}
		recordList = append(recordList, record)
	}

	count := 0
	var specList []pcData.GPUSpec
	go func() {
		for {
			<-ticker.C
			gpuRecord := pcData.GetGPUSpec(recordList[count])
			specList = append(specList, gpuRecord)
			count++
			if count == len(recordList) {
				saveSpecData(specList, "gpu")
				ticker.Stop()
				runtime.Goexit()
			}
		}
	}()

	listLen := time.Duration(timeSet * (len(recordList) + 3))
	time.Sleep(time.Second * listLen)
}

/*
RAM DATA
*/
func updateSpecLogic(name string) {
	timeSet := 3000
	timeDuration := time.Duration(timeSet) * time.Millisecond
	ticker := time.NewTicker(timeDuration)

	dataList := readCsvFile("res/" + name + "data.csv")
	var recordList []pcData.LinkRecord

	for i := 1; i < len(dataList); i++ {
		data := dataList[i]
		record := pcData.LinkRecord{Brand: data[0], Name: data[1], PriceCN: data[2], LinkSpec: data[3], LinkCN: data[4], LinkUS: data[5], LinkHK: data[6]}
		recordList = append(recordList, record)
	}

	count := 0
	switch name {
	case "cpu":
		var specList []pcData.CPUSpec
		go func() {
			for {
				<-ticker.C
				cpuRecord := pcData.GetCPUSpec(recordList[count])
				specList = append(specList, cpuRecord)
				count++
				if count == len(recordList) {
					saveSpecData(specList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()
	case "mb":
		var specList []pcData.MotherboardSpec
		go func() {
			for {
				<-ticker.C
				mbRecord := pcData.GetMotherboardSpec(recordList[count])
				specList = append(specList, mbRecord)
				count++
				if count == len(recordList) {
					saveSpecData(specList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()
	case "ram":
		var specList []pcData.RamSpec
		go func() {
			for {
				<-ticker.C
				ramRecord := pcData.GetRamSpec(recordList[count])
				specList = append(specList, ramRecord)
				count++
				if count == 1 {
					saveSpecData(specList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()
	default:
		var specList []pcData.RamSpec
		go func() {
			for {
				<-ticker.C
				ramRecord := pcData.GetRamSpec(recordList[count])
				specList = append(specList, ramRecord)
				count++
				if count == 1 {
					saveSpecData(specList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()
	}

	listLen := time.Duration(timeSet * (len(recordList) + 3))
	time.Sleep(time.Second * listLen)
}

func updatePriceLogic(name string) {
	specFile, _ := os.Open("tmp/spec/" + name + "Spec.json")
	byteValue, _ := io.ReadAll(specFile)

	switch name {
	case "cpu":
	case "ram":
		var specList []pcData.RamSpec
		var ramList []pcData.RamType

		json.Unmarshal([]byte(byteValue), &specList)

		for i := 0; i < len(specList); i++ {
			spec := specList[i]
			record := pcData.GetRamData(spec)
			ramList = append(ramList, record)

			if i == 1 {
				saveData(ramList, name+"Data")
			}
		}

	default:
		var specList []pcData.RamSpec
		var ramList []pcData.RamType

		json.Unmarshal([]byte(byteValue), &specList)

		for i := 1; i < len(specList); i++ {
			spec := specList[i]
			record := pcData.GetRamData(spec)
			ramList = append(ramList, record)

			if i == 1 {
				saveData(ramList, name+"Data")
			}
		}
	}
}
