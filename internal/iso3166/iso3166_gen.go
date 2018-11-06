// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

// +build ignore

// Generates iso3166.go.
//
// This file grabs the ISO 3166-1-alpha2 codes and writes them
// into source code so we don't rely on any external files (zip,
// json, etc).
//
// The data is pulled from datahub.io as the ISO.org site only offers
// XML.
//
// https://datahub.io/core/country-list#data
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"net/http"
	"os/user"
	"runtime"
	"time"
)

const (
	// From https://datahub.io/core/country-list#data
	downloadUrl = "https://datahub.io/core/country-list/r/data.json"

	outputFilename = "internal/iso3166/iso3166.go"
)

// [{"Code": "AF", "Name": "Afghanistan"}, ...]
type country struct {
	Code string `json:"Code"`
	Name string `json:"Name"`
}

func main() {
	when := time.Now().Format("2006-01-02T03:04:05Z")
	who, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to get user on %s", runtime.GOOS)
	}

	// Write copyright header
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

// Generated on %s by %s, any modifications will be overwritten
package iso3166
`, when, who.Username)

	// Download certs
	resp, err := http.Get(downloadUrl)
	if err != nil {
		log.Fatalf("error while downloading %s: %v", downloadUrl, err)
	}
	defer resp.Body.Close()

	var countries []country
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		log.Fatalf("error while parsing country response: %v", err)
	}

	// Write countries to source code
	fmt.Fprintln(&buf, "var countryCodes = map[string]bool{")
	for i := range countries {
		fmt.Fprintf(&buf, fmt.Sprintf(`"%s": true, // %s`+"\n", countries[i].Code, countries[i].Name))
	}
	fmt.Fprintln(&buf, "}")

	// format source code and write file
	out, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println(buf.String())
		log.Fatalf("error formatting output code, err=%v", err)
	}

	err = ioutil.WriteFile(outputFilename, out, 0644)
	if err != nil {
		log.Fatalf("error writing file, err=%v", err)
	}
}