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
	// pcData.GetRamCNPriceFromChromedp("https://item.taobao.com/item.htm?abbucket=17&id=743688559462&skuId=5323436787436")

	const (
		cpu         = "cpu"
		gpu         = "gpu"
		motherboard = "motherboard"
		ram         = "ram"
		ssd         = "ssd"
		power       = "power"
		cooler      = "cooler"
		pcCase      = "case"
	)

	getDataName := cpu
	isUpdateSpec := false

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
	timeSet := 5000
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

func updateSpecLogic(name string) {
	timeSet := 5000
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
	case "motherboard":
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
				if count == len(recordList) {
					saveSpecData(specList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()
	case "ssd":
		var specList []pcData.SSDSpec
		go func() {
			for {
				<-ticker.C
				ssdRecord := pcData.GetSSDSpec(recordList[count])
				specList = append(specList, ssdRecord)
				count++
				if count == len(recordList) {
					saveSpecData(specList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()
	case "power":
		var specList []pcData.PowerSpec
		go func() {
			for {
				<-ticker.C
				powerRecord := pcData.GetPowerSpec(recordList[count])
				specList = append(specList, powerRecord)
				count++
				if count == len(recordList) {
					saveSpecData(specList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()
	case "cooler":
		var specList []pcData.CoolerSpec
		go func() {
			for {
				<-ticker.C
				coolerRecord := pcData.GetCoolerSpec(recordList[count])
				specList = append(specList, coolerRecord)
				count++
				if count == len(recordList) {
					saveSpecData(specList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()
	default:
		var specList []pcData.CaseSpec
		go func() {
			for {
				<-ticker.C
				caseRecord := pcData.GetCaseSpec(recordList[count])
				specList = append(specList, caseRecord)
				count++
				if count == len(recordList) {
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
	timeSet := 5000
	timeDuration := time.Duration(timeSet) * time.Millisecond
	ticker := time.NewTicker(timeDuration)

	specFile, _ := os.Open("tmp/spec/" + name + "Spec.json")
	byteValue, _ := io.ReadAll(specFile)
	count := 0

	switch name {
	case "cpu":
		var specList []pcData.CPUSpec
		var cpuList []pcData.CPUType
		json.Unmarshal([]byte(byteValue), &specList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record := pcData.GetCPUData(spec)
				cpuList = append(cpuList, record)

				count++
				if count == len(specList) {
					saveData(cpuList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + 3))
		time.Sleep(time.Second * listLen)
	case "gpu":
		var specList []pcData.GPUSpec
		var gpuList []pcData.GPUType

		json.Unmarshal([]byte(byteValue), &specList)

		dataList := readCsvFile("res/" + name + "data.csv")
		var recordList []pcData.GPURecordData

		for i := 1; i < len(dataList); i++ {
			data := dataList[i]
			record := pcData.GPURecordData{Brand: data[0], Name: data[1], PriceCN: data[2], LinkCN: data[3], LinkUS: data[4], LinkHK: data[5]}
			recordList = append(recordList, record)
		}

		go func() {
			for {
				<-ticker.C
				data := recordList[count]
				record := pcData.GetGPUData(specList, data)
				gpuList = append(gpuList, record)

				if count == len(recordList) {
					saveData(gpuList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(recordList) + 3))
		time.Sleep(time.Second * listLen)
	case "ram":
		var specList []pcData.RamSpec
		var ramList []pcData.RamType

		json.Unmarshal([]byte(byteValue), &specList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record := pcData.GetRamData(spec)
				ramList = append(ramList, record)

				if count == len(specList) {
					saveData(ramList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + 3))
		time.Sleep(time.Second * listLen)
	case "ssd":
		var specList []pcData.SSDSpec
		var ssdList []pcData.SSDType

		json.Unmarshal([]byte(byteValue), &specList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record := pcData.GetSSDData(spec)
				ssdList = append(ssdList, record)

				if count == len(specList) {
					saveData(ssdList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + 3))
		time.Sleep(time.Second * listLen)
	default:
		var specList []pcData.RamSpec
		var ramList []pcData.RamType

		json.Unmarshal([]byte(byteValue), &specList)

		go func() {
			for {
				<-ticker.C
				spec := specList[count]
				record := pcData.GetRamData(spec)
				ramList = append(ramList, record)

				if count == len(specList) {
					saveData(ramList, name)
					ticker.Stop()
					runtime.Goexit()
				}
			}
		}()

		listLen := time.Duration(timeSet * (len(specList) + 3))
		time.Sleep(time.Second * listLen)
	}
}
