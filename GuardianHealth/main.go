package main

import (
	database "GuardianHealth/database_Conn"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Patient represents a patient's record
type Patient struct {
	ID         string
	Name       string
	Age        int
	Diagnosis  string
	Metadata   map[string]interface{} // Metadata fields
	Taxonomies []string               // Taxonomy tags
}

func main() {
	// Create a new instance of RedisDB with an encryption key
	redisDB := database.NewRedisDB("encryptionkeys12")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n1. Store Patient Data")
		fmt.Println("2. Retrieve Patient Data")
		fmt.Println("3. Delete All Data")
		fmt.Println("4. Exit")
		fmt.Print("Enter your choice: ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading choice:", err)
			continue
		}
		choice = strings.TrimSpace(choice) // Trim newline character

		// Clear input buffer
		_, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error clearing input buffer:", err)
			continue
		}

		switch choice {
		case "1":
			fmt.Println("Enter Patient ID:")
			id, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading ID:", err)
				continue
			}
			id = strings.TrimSpace(id) // Trim newline character

			fmt.Println("Enter Patient Name:")
			name, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading name:", err)
				continue
			}
			name = strings.TrimSpace(name) // Trim newline character

			fmt.Println("Enter Patient Diagnosis:")
			diagnosis, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading diagnosis:", err)
				continue
			}
			diagnosis = strings.TrimSpace(diagnosis) // Trim newline character

			fmt.Println("Enter Patient Age:")
			var age int
			_, err = fmt.Scan(&age)
			if err != nil {
				fmt.Println("Error reading age:", err)
				continue
			}

			// Read metadata input
			metadata, err := readMetadata(reader)
			if err != nil {
				fmt.Println("Error reading metadata:", err)
				continue
			}

			// Read taxonomy input
			fmt.Println("Enter Patient Taxonomies (comma-separated):")
			taxonomiesStr, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading taxonomies:", err)
				continue
			}
			taxonomies := strings.Split(strings.TrimSpace(taxonomiesStr), ",")

			patient := Patient{
				ID:         id,
				Name:       name,
				Age:        age,
				Diagnosis:  diagnosis,
				Metadata:   metadata,
				Taxonomies: taxonomies,
			}
			err = redisDB.StorePatient(patient.ID, patient)
			if err != nil {
				fmt.Println("Error storing patient data:", err)
			} else {
				fmt.Println("Patient data stored successfully!")
			}

		case "2":
			fmt.Println("Enter Patient ID:")
			id, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading ID:", err)
				continue
			}
			id = strings.TrimSpace(id) // Trim newline character

			var retrievedPatient Patient
			err = redisDB.GetPatient(id, &retrievedPatient)
			if err != nil {
				fmt.Println("Error retrieving patient data:", err)
			} else {
				fmt.Println("Retrieved Patient Data:")
				fmt.Println("ID:", retrievedPatient.ID)
				fmt.Println("Name:", retrievedPatient.Name)
				fmt.Println("Age:", retrievedPatient.Age)
				fmt.Println("Diagnosis:", retrievedPatient.Diagnosis)
				fmt.Println("Metadata:", retrievedPatient.Metadata)
				fmt.Println("Taxonomies:", retrievedPatient.Taxonomies)
			}

		case "3":
			// Delete All Data
			fmt.Println("Are you sure you want to delete all data? (yes/no):")
			confirmation, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading confirmation:", err)
				continue
			}
			confirmation = strings.TrimSpace(confirmation) // Trim newline character
			if confirmation == "yes" {
				err := redisDB.FlushAll()
				if err != nil {
					fmt.Println("Error deleting all data:", err)
				} else {
					fmt.Println("All data deleted successfully!")
				}
			} else {
				fmt.Println("Operation cancelled.")
			}
		case "4":
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Invalid choice. Please enter a valid option.")
		}
	}
}

func readMetadata(reader *bufio.Reader) (map[string]interface{}, error) {
	fmt.Println("Enter Patient Metadata (in JSON format):")
	metadataStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading metadata: %w", err)
	}
	metadataStr = strings.TrimSpace(metadataStr) // Trim newline character

	if metadataStr == "" {
		return nil, nil // No metadata provided
	}

	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(metadataStr), &metadata)
	if err != nil {
		return nil, fmt.Errorf("error parsing metadata: %w", err)
	}

	return metadata, nil
}
