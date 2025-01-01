package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// create const for api paths
// on get of api path, maybe ask for header args, and serve the associated json
var paths = map[string]string{
	"staffList": "./data/staff/staffList.json",
}

func main() {
	handleRequests()
}

func getJson(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	return bytes, nil
}

func appendJson(filePath string, newData interface{}) error {
	// Open the file for reading and writing
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Read the file content into a byte slice
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Parse the existing JSON into a slice (root array)
	var existingData []map[string]interface{} // Assuming we're appending to an array of objects
	if len(data) > 0 {
		err = json.Unmarshal(data, &existingData)
		if err != nil {
			return fmt.Errorf("error unmarshalling existing data: %v", err)
		}
	}

	// Append the new data (which should be a single staff member object)
	existingData = append(existingData, newData.(map[string]interface{})) // Assuming newData is a map

	// Marshal the updated data back into JSON
	updatedData, err := json.MarshalIndent(existingData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling updated data: %v", err)
	}

	// Write the updated data back to the file
	err = os.WriteFile(filePath, updatedData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

func staffList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, err := getJson(paths["staffList"])
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read staff list: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	//_ = TAKE THESE BYTES BITCH I WILL NOT TELL YOU HOW MANY THERE ARE
	_, writeErr := w.Write(data)
	if writeErr != nil {
		fmt.Println("Failed to write response: ", writeErr)
	}
}

func addStaff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Step 1: Decode the incoming JSON object (a single staff member)
	var newStaff map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newStaff)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Step 2: Call appendJson to append the new staff data
	err = appendJson(paths["staffList"], newStaff)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error appending staff data: %v", err), http.StatusInternalServerError)
		return
	}

	// Step 3: Return a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Staff data added successfully"}`))
}

func handleRequests() {
	http.HandleFunc("/staff/staffList", staffList)
	http.HandleFunc("/staff/add", addStaff)

	log.Fatal(http.ListenAndServe(":8080", nil))
	//Change nil to be a handler function, likely serve the svelte 404
}
