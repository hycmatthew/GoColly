package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func saveTranslationData(result any, name string) {
	fmt.Println("save translation started!")
	jsonData, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write JSON data to file
	err = os.WriteFile("res/"+name+"/translation.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func main() {
	csvFile, err := os.Open("../res/translation/pc-part-translation.csv")
	if err != nil {
		fmt.Printf("Error opening CSV file: %v\n", err)
		return
	}
	defer csvFile.Close()

	// Create a CSV reader
	reader := csv.NewReader(csvFile)

	// Read the headers (first row)
	headers, err := reader.Read()
	if err != nil {
		fmt.Printf("Error reading headers: %v\n", err)
		return
	}

	// Prepare a map to hold data for each header
	var headerList []string
	dataMap := make(map[string][]string)

	// Read the rest of the rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break // End of file
		}
		if err != nil {
			fmt.Printf("Error reading row: %v\n", err)
			return
		}

		// Populate the map with rows grouped by header
		for i, value := range row {
			header := headers[i]
			headerList = append(headerList, value)
			dataMap[header] = append(dataMap[header], value)
		}
	}

	for header, value := range dataMap {
		if header != "id" {
			data := make(map[string]string)
			for i, key := range dataMap["id"] {
				data[key] = value[i]
			}
			saveTranslationData(data, header)
		}
	}

	// Write each header's data to a separate JSON file
	/*
		for header, values := range dataMap {
			// Create a JSON file for this header
			filePath := filepath.Join(outputDir, header+".json")
			file, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("Error creating JSON file for header '%s': %v\n", header, err)
				return
			}
			defer file.Close()

			// Write the JSON data
			jsonData, err := json.MarshalIndent(values, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling JSON for header '%s': %v\n", header, err)
				return
			}
			_, err = file.Write(jsonData)
			if err != nil {
				fmt.Printf("Error writing JSON for header '%s': %v\n", header, err)
				return
			}

			fmt.Printf("JSON file created for header: %s\n", header)
		}
	*/
}
