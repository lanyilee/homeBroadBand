package core

import (
	"encoding/csv"
	"os"
)

type CsvUtil struct {
}

func (csvUtil *CsvUtil) Put(data string) error {
	file, err := os.OpenFile("./date.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	defer file.Close()
	if err != nil {
		Logger("open date.csv error")
		return err
	}
	writter := csv.NewWriter(file)
	defer writter.Flush()
	err = writter.Write([]string{data})
	if err != nil {
		Logger("write data to date.csv error")
		return err
	}
	return nil
}

func (csvUtil *CsvUtil) IsExist(data string) (b bool, err error) {
	b = false
	file, err := os.OpenFile("./date.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	defer file.Close()
	if err != nil {
		Logger("open date.csv error")
		return true, err
	}
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 0
	record, err := reader.ReadAll()
	if err != nil {
		Logger("read date.csv error")
		return true, err
	}
	for _, item := range record {
		if item[0] == data {
			b = true
			break
		}
	}
	return b, nil
}
