package main

import (
	"crypto/sha512"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type city struct {
	URL          string
	Output       string
	IgnoreHashes []string
}

func init() {
	flag.Usage = usage
}

// TODO: Write unit tests :)
func main() {
	executableName := os.Args[0]
	var configFile string
	var createConfigSample bool
	var shouldLogHashCheck bool
	var ignoreWriteErrors bool

	flag.StringVar(&configFile, "config-file", executableName+".json", "The json configuration file.")
	flag.BoolVar(&createConfigSample, "create-config-file", false, fmt.Sprintf("Create the json configuration file %v.json if it does not exists and then exit.", executableName))
	flag.BoolVar(&ignoreWriteErrors, "ignore-errors", false, "Ignore errors in fetching or writing files.")
	flag.BoolVar(&shouldLogHashCheck, "log-hash-check", false, "Log the hashes that are not in ignore list.")

	cities, err := readConfig(configFile)
	if err != nil {
		fmt.Printf("Cannot read config file: %s \nerror : %v\n\n", configFile, err)
		flag.Usage()
		os.Exit(3)
	}

	client := http.Client{}
	anyError := false
	for _, selectedCity := range cities {
		image, err := getWeatherData(client, selectedCity.URL)
		if err != nil {
			log.Println("Error calling url ", err)
		} else {
			if checkHashDoesNotMatch(image, selectedCity.IgnoreHashes, shouldLogHashCheck) {
				if err := ioutil.WriteFile(selectedCity.Output, image, os.FileMode(0644)); err != nil {
					fmt.Println("Could not write file", selectedCity.Output, err)
					anyError = true
				} else {
					fmt.Println("Wrote file ", selectedCity.Output)
				}
			}
		}
	}

	if anyError {
		os.Exit(1)
	}
}

func checkHashDoesNotMatch(image []byte, hashes []string, shouldLog bool) bool {
	imageHashBytes := sha512.Sum512(image)
	for _, hash := range hashes {
		imageHash := fmt.Sprintf("%x", imageHashBytes)
		if imageHash == hash {
			return false
		}
		if shouldLog {
			log.Printf("Hash not in ignore list : %v", imageHash)
		}
	}
	return true
}

func getWeatherData(client http.Client, url string) ([]byte, error) {
	var empty []byte
	response, err := client.Get(url)
	if err != nil {
		return empty, err
	}

	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return empty, err
		}

		return body, nil
	}

	return empty, fmt.Errorf("Url returned %v", response.StatusCode)
}

func readConfig(configFile string) ([]city, error) {
	var cities []city
	var empty []city
	dat, err := ioutil.ReadFile(configFile)
	if err != nil {
		return empty, err
	}

	err = json.Unmarshal(dat, &cities)
	if err != nil {
		return empty, err
	}

	return cities, nil
}

// TODO : write sample config here
func writeConfigFile() {
	fmt.Println("Will write sample config with wttr.in/pune_0tqp0.png")
}

func usage() {
	executableName := os.Args[0]
	fmt.Fprintf(flag.CommandLine.Output(), "\n%s is a wttr weather retriever.\n", executableName)
	fmt.Fprintf(flag.CommandLine.Output(), "Below are the Optional arguments : \n")
	flag.PrintDefaults()
	os.Exit(3)
}
