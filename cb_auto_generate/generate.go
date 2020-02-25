package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

//go:generate go run generate.go
func main() {
	files := []struct {
		FileName           string
		FilePath           string
		OutputFilePath     string
		OutputPackageName  string
		OutputFunctionName string
	}{
		{
			FileName:           "master.gohtml",
			FilePath:           "../module/cbwebcommon/master.gohtml",
			OutputFilePath:     "../module/cbwebcommon/globalmastertemplate.go",
			OutputPackageName:  "cbwebcommon",
			OutputFunctionName: "getGlobalMasterTemplate",
		},
		{
			FileName:           "nav.gohtml",
			FilePath:           "../module/cbwebcommon/nav.gohtml",
			OutputFilePath:     "../module/cbwebcommon/globalnavtemplate.go",
			OutputPackageName:  "cbwebcommon",
			OutputFunctionName: "getGlobalNavTemplate",
		},
		{
			FileName:           "flash.gohtml",
			FilePath:           "../module/cbwebcommon/flash.gohtml",
			OutputFilePath:     "../module/cbwebcommon/globalflashtemplate.go",
			OutputPackageName:  "cbwebcommon",
			OutputFunctionName: "getGlobalFlashTemplate",
		},
		{
			FileName:           "inputtext.gohtml",
			FilePath:           "../module/cbwebcommon/inputtext.gohtml",
			OutputFilePath:     "../module/cbwebcommon/globalinputtexttemplate.go",
			OutputPackageName:  "cbwebcommon",
			OutputFunctionName: "getGlobalInputTextTemplate",
		},
		{
			FileName:           "inputselect.gohtml",
			FilePath:           "../module/cbwebcommon/inputselect.gohtml",
			OutputFilePath:     "../module/cbwebcommon/globalinputselecttemplate.go",
			OutputPackageName:  "cbwebcommon",
			OutputFunctionName: "getGlobalInputSelectTemplate",
		},
		{
			FileName:           "datatable.gohtml",
			FilePath:           "../module/cbwebcommon/datatable.gohtml",
			OutputFilePath:     "../module/cbwebcommon/globaldatatabletemplate.go",
			OutputPackageName:  "cbwebcommon",
			OutputFunctionName: "getGlobalDataTableTemplate",
		},
		{
			FileName:           "flashtoast.gohtml",
			FilePath:           "../module/cbwebcommon/flashtoast.gohtml",
			OutputFilePath:     "../module/cbwebcommon/globalflashtoasttemplate.go",
			OutputPackageName:  "cbwebcommon",
			OutputFunctionName: "getGlobalFlashToastTemplate",
		},
	}

	for _, file := range files {
		fileBytes, e := ioutil.ReadFile(file.FilePath)
		if e != nil {
			panic(e)
		}

		var bytesString []string

		for _, fileByte := range fileBytes {
			bytesString = append(bytesString, strconv.Itoa(int(fileByte)))
		}

		e = ioutil.WriteFile(file.OutputFilePath, []byte(stringReplacer(`package {packageName}

// DO NOT EDIT: This is autogenerated from {fileName}
// run go generate in the cb_auto_generate directory to regenerate this
func {functionName}() []byte {
	return []byte{{byteArrayString}}
}
`,
			"{packageName}", file.OutputPackageName,
			"{fileName}", file.FileName,
			"{functionName}", file.OutputFunctionName,
			"{byteArrayString}", strings.Join(bytesString, ","),
		)), os.ModePerm)
		if e != nil {
			panic(e)
		}
	}
}

func stringReplacer(format string, args ...string) string {
	r := strings.NewReplacer(args...)
	return r.Replace(format)
}
