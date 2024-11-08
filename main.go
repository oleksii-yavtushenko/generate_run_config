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

	flag.Parse()

	if *workingDir == "" || *moduleName == "" || *packageValue == "" {
		fmt.Println("All flags are required: -workingDir, -moduleName, -package")
		os.Exit(1)
	}

	lastDirsNum := 2
	configName := getConfigName(*workingDir, lastDirsNum)
	// Generate the XML file name by replacing slashes with underscores
	fileName := generateFileName(configName)

	fullDirPath := prefixProjectDir(*workingDir)

	packagePath := generatePackagePath(*packageValue, *moduleName, *workingDir)

	folderName := extractFolderName(*workingDir)

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

	// Ensure the .idea/runConfigurations directory exists
	runConfigDir := filepath.Join(".", ".idea", "runConfigurations")
	err = os.MkdirAll(runConfigDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	// Write the configuration to a file
	runConfigFile := filepath.Join(runConfigDir, fileName+".xml")
	err = os.WriteFile(runConfigFile, output, 0644)
	if err != nil {
		fmt.Println("Error writing run configuration file:", err)
		return
	}

	fmt.Println("Run configuration generated successfully at", runConfigFile)
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
