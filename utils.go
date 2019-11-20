/*
Drone plugin to upload one or more packages to Bintray.
See README.md for usage.
Author: Archit Sharma December 2019 (Github arcolife)
Previous: David Tootill November 2015 (GitHub tooda02)
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// PrettyPrint to print a struct like JSON
func PrettyPrint(v interface{}, label string) (err error) {
	fmt.Println(label, "=>")
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

// ReadConfig to UnMarshall tunables and settings from YAML file
func (config *BintrayConfig) ReadConfig(cfgPath string) {
	configdata, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		fmt.Print(errors.Wrap(err, "Error reading file"))
		os.Exit(1)
	}

	err = yaml.Unmarshal(configdata, &config)
	if err != nil {
		fmt.Print(errors.Wrap(err, "Error Unmarshalling file"))
		os.Exit(1)
	}
}

func cleanup() {
	fmt.Printf("\nAborting. Caught a signal fron os.Interrupt\n")
}
