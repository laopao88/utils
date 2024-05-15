package zaia

import (
	"os"
	"text/template"
)

func Parse(fileName string, outFileName string, mapInfo map[string]interface{}) error {
	t, err := template.ParseFiles(fileName)
	if err != nil {
		return err
	}
	f, err := os.Create(outFileName)
	defer f.Close()
	if err != nil {
		return err
	}
	err = t.Execute(f, mapInfo)
	return err
}
