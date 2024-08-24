package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"go-colly-lib/src/pcData"
	"log"
	"os"
	"runtime"
	"time"
)

func main() {
	getGPUSpecLogic()
	/*
		getDataNum := "spec"
		if getDataNum == "cpu" {
			udpateCPULogic()
		} else if getDataNum == "spec" {
			getGPUSpecLogic()
		} else if getDataNum == "gpu" {
			udpateGPULogic()
		}
	*/
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

func udpateCPULogic() {
	dataList := readCsvFile("res/cpudata.csv")
	var recordList []pcData.CPURecord
	var cpuList []pcData.CPUType

	for i := 1; i < len(dataList); i++ {
		data := dataList[i]
		record := pcData.CPURecord{Name: data[0], LinkSpec: data[1], LinkCN: data[2], LinkUS: data[3], LinkHK: data[4]}
		recordList = append(recordList, record)
	}

	ticker := time.NewTicker(1500 * time.Millisecond)
	count := 0

	go func() {
		for {
			<-ticker.C

			cpuRecord := pcData.GetCPUData(recordList[count].LinkSpec, recordList[count].LinkUS, recordList[count].LinkCN, recordList[count].LinkHK)
			cpuList = append(cpuList, cpuRecord)
			count++
			if count == len(recordList) {
				saveData(cpuList, "cpuData")
				ticker.Stop()
				runtime.Goexit()
			}
		}
	}()

	listLen := time.Duration((len(recordList) * 2) + 3)
	time.Sleep(time.Second * listLen)
}

func getGPUSpecLogic() {
	timeSet := 4000
	timeDuration := time.Duration(timeSet) * time.Millisecond
	ticker := time.NewTicker(timeDuration)

	specdataList := pcData.GetGPUSpecDataList()
	var specList []pcData.GPUSpecTempStruct

	count := 0

	go func() {
		for {
			<-ticker.C

			spec := pcData.GetGPUSpec(specdataList[count].Name, specdataList[count].Link)
			specList = append(specList, spec)
			count++

			if count == len(specdataList) {
				saveData(specList, "gpuSpec")
				ticker.Stop()
				runtime.Goexit()
			}
		}
	}()

	listLen := time.Duration(timeSet * (len(specdataList) + 2))
	time.Sleep(listLen * time.Millisecond)
}

func udpateGPULogic() {
	specdataList := pcData.GetGPUSpecDataList()
	dataList := readCsvFile("res/gpudata.csv")
	var recordList []pcData.GPURecord
	var specList []pcData.GPUSpecTempStruct
	var gpuList []pcData.GPUType

	for i := 1; i < len(dataList); i++ {
		data := dataList[i]
		record := pcData.GPURecord{Name: data[1], Brand: data[0], LinkCN: data[2], LinkUS: data[3], LinkHK: data[4]}
		recordList = append(recordList, record)
	}

	ticker := time.NewTicker(3500 * time.Millisecond)
	count := 0
	totalLen := len(specdataList) + len(recordList)

	go func() {
		for {
			<-ticker.C

			if count < len(specdataList) {
				data := specdataList[count]
				spec := pcData.GetGPUSpec(data.Name, data.Link)
				specList = append(specList, spec)
			} else {
				newCount := count - len(specdataList)
				tempData := recordList[newCount]
				gpuRecord := pcData.GetGPUData(specList, tempData.Brand, tempData.LinkUS, tempData.LinkCN, tempData.LinkHK)
				gpuList = append(gpuList, gpuRecord)
			}

			count++
			if count == totalLen {
				saveData(gpuList, "gpuData")
				ticker.Stop()
				runtime.Goexit()
			}
		}
	}()

	listLen := time.Duration((len(recordList) * 2) + 3)
	time.Sleep(time.Second * listLen)
}
