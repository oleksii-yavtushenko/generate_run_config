package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	Separator  = string(os.PathSeparator)
	ProjectDir = "$PROJECT_DIR$"
)

type Component struct {
	XMLName       xml.Name      `xml:"component"`
	Name          string        `xml:"name,attr"`
	Configuration Configuration `xml:"configuration"`
}

type Configuration struct {
	XMLName          xml.Name   `xml:"configuration"`
	Default          bool       `xml:"default,attr"`
	Name             string     `xml:"name,attr"`
	Type             string     `xml:"type,attr"`
	FactoryName      string     `xml:"factoryName,attr"`
	FolderName       string     `xml:"folderName,attr"`
	Module           Module     `xml:"module"`
	WorkingDirectory WorkingDir `xml:"working_directory"`
	Kind             Kind       `xml:"kind"`
	Package          Package    `xml:"package"`
	FilePath         FilePath   `xml:"filePath"`
	Method           Method     `xml:"method"`
	Directory        Directory  `xml:"directory"`
}

type Module struct {
	Name string `xml:"name,attr"`
}

type WorkingDir struct {
	Value string `xml:"value,attr"`
}

type Kind struct {
	Value string `xml:"value,attr"`
}

type Package struct {
	Value string `xml:"value,attr"`
}

type FilePath struct {
	Value string `xml:"value,attr"`
}

type Method struct {
	V string `xml:"v,attr"`
}

type Directory struct {
	Value string `xml:"value,attr"`
}

func main() {

	// Flags to specify directory and other settings
	workingDir := flag.String("workingDir", "", "Working directory (required)")
	moduleName := flag.String("moduleName", "", "Module name (required)")
	packageValue := flag.String("package", "", "Go package (required)")
	current := flag.Bool("current", false, "Generate run configuration in current directory and remove existing ones")

	flag.Parse()

	if *workingDir == "" || *moduleName == "" || *packageValue == "" {
		fmt.Println("All flags are required: -workingDir, -moduleName, -package")
		os.Exit(1)
	}

	lastDirsNum := 2
	configName := getConfigName(*workingDir, lastDirsNum)

	// Generate the XML file name - add "current_" prefix if the current flag is set
	var fileName string
	if *current {
		fileName = "current_" + generateFileName(configName)
	} else {
		fileName = generateFileName(configName)
	}

	fullDirPath := prefixProjectDir(*workingDir)

	packagePath := generatePackagePath(*packageValue, *moduleName, *workingDir)

	// Set folder name based on current flag
	var folderName string
	if *current {
		folderName = "current"
	} else {
		folderName = extractFolderName(*workingDir)
	}

	// Create the configuration structure
	runConfig := Component{
		Name: "ProjectRunConfigurationManager",
		Configuration: Configuration{
			Default:     false,
			Name:        configName,
			Type:        "GoApplicationRunConfiguration",
			FactoryName: "Go Application",
			FolderName:  folderName,
			Module: Module{
				Name: *moduleName,
			},
			WorkingDirectory: WorkingDir{
				Value: fullDirPath,
			},
			Kind: Kind{
				Value: "PACKAGE",
			},
			Package: Package{
				Value: packagePath,
			},
			FilePath: FilePath{
				Value: fullDirPath,
			},
			Method: Method{
				V: "2",
			},
			Directory: Directory{
				Value: ProjectDir,
			},
		},
	}

	// Marshal the XML
	output, err := xml.MarshalIndent(runConfig, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling XML:", err)
		return
	}

	// Always use .idea/runConfigurations directory
	runConfigDir := filepath.Join(".", ".idea", "runConfigurations")
	err = os.MkdirAll(runConfigDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	// If current flag is set, remove existing configurations with "current_" prefix
	if *current {
		err = removeCurrentConfigurations(runConfigDir)
		if err != nil {
			fmt.Println("Error removing existing current configurations:", err)
			return
		}
	}

	// Write the configuration to a file
	runConfigFile := filepath.Join(runConfigDir, fileName+".xml")

	// Add XML header
	xmlHeader := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fullOutput := append(xmlHeader, output...)

	err = os.WriteFile(runConfigFile, fullOutput, 0644)
	if err != nil {
		fmt.Println("Error writing run configuration file:", err)
		return
	}

	// Get absolute path for better output
	absPath, err := filepath.Abs(runConfigFile)
	if err != nil {
		absPath = runConfigFile
	}

	fmt.Println("Run configuration generated successfully at", absPath)
}

// removeCurrentConfigurations removes all XML files that start with "current_" and look like run configurations
func removeCurrentConfigurations(dir string) error {
	// Read all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var removedCount int
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		// Check if it's an XML file that starts with "current_"
		if strings.HasSuffix(fileName, ".xml") && strings.HasPrefix(fileName, "current_") {
			filePath := filepath.Join(dir, fileName)

			// Read the file to check if it's a run configuration
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue // Skip files we can't read
			}

			// Check if it contains run configuration markers
			contentStr := string(content)
			if strings.Contains(contentStr, "ProjectRunConfigurationManager") ||
				strings.Contains(contentStr, "GoApplicationRunConfiguration") ||
				strings.Contains(contentStr, "component name=\"ProjectRunConfigurationManager\"") {

				err = os.Remove(filePath)
				if err != nil {
					fmt.Printf("Warning: failed to remove %s: %v\n", filePath, err)
				} else {
					fmt.Printf("Removed existing current configuration: %s\n", fileName)
					removedCount++
				}
			}
		}
	}

	if removedCount > 0 {
		fmt.Printf("Removed %d existing current run configuration(s)\n", removedCount)
	}

	return nil
}

// removeExistingConfigurations removes all XML files that look like run configurations from the specified directory
func removeExistingConfigurations(dir string) error {
	// Read all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var removedCount int
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		// Check if it's an XML file that might be a run configuration
		if strings.HasSuffix(fileName, ".xml") {
			filePath := filepath.Join(dir, fileName)

			// Read the file to check if it's a run configuration
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue // Skip files we can't read
			}

			// Check if it contains run configuration markers
			contentStr := string(content)
			if strings.Contains(contentStr, "ProjectRunConfigurationManager") ||
				strings.Contains(contentStr, "GoApplicationRunConfiguration") ||
				strings.Contains(contentStr, "component name=\"ProjectRunConfigurationManager\"") {

				err = os.Remove(filePath)
				if err != nil {
					fmt.Printf("Warning: failed to remove %s: %v\n", filePath, err)
				} else {
					fmt.Printf("Removed existing configuration: %s\n", fileName)
					removedCount++
				}
			}
		}
	}

	if removedCount > 0 {
		fmt.Printf("Removed %d existing run configuration(s)\n", removedCount)
	}

	return nil
}

// getConfigName generates the configuration name based on the folder structure of the file path
func getConfigName(path string, lastDirsNum int) string {
	// Convert the directory structure to a configuration name
	dirArr := strings.Split(path, Separator)
	// Last two dirs
	return strings.Join(dirArr[len(dirArr)-lastDirsNum:], Separator)
}

// generateFileName generates the XML file name by replacing slashes with underscores
func generateFileName(configName string) string {
	// Replace slashes with underscores
	return strings.ReplaceAll(configName, Separator, "_")
}

// prefixProjectDir adds prefix project dir to the provided directory
func prefixProjectDir(dir string) string {
	return fmt.Sprintf("%s%s%s", ProjectDir, Separator, dir)
}

// generatePackagePath generates absolute package paths from relative
func generatePackagePath(packageValue, module, workingDir string) string {
	// Trim separator
	packageValue = strings.TrimSuffix(packageValue, Separator)

	// Join with separator
	return strings.Join([]string{packageValue, module, workingDir}, Separator)
}

func extractFolderName(path string) string {
	// Convert the directory structure to a configuration name
	dirArr := strings.Split(path, Separator)
	// 3rd dir before end
	return dirArr[len(dirArr)-3]
}

func isEndpoint(path string) bool {
	restMethods := []string{"get", "patch", "post", "put", "delete"}
	name := getConfigName(path, 2)
	for _, method := range restMethods {
		if strings.Contains(name, method) {
			return true
		}
	}
	return false
}
