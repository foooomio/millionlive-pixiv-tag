package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

const baseUrl = "https://dic.pixiv.net/a/"

var re = regexp.MustCompile(`pixivに投稿された作品数: (\d+)`)

func readCsv(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func writeCsv(path string, records [][]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.WriteAll(records)
	if err != nil {
		return err
	}

	return nil
}

func fetch(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func fetchTagCount(name string) (string, error) {
	data, err := fetch(baseUrl + name)
	if err != nil {
		return "", err
	}

	match := re.FindSubmatch(data)
	if match == nil {
		return "", fmt.Errorf("not match: %s", name)
	}
	return string(match[1]), nil
}

func update(records [][]string) ([][]string, error) {
	output := make([][]string, len(records))

	today := time.Now().Format("2006-01-02")
	output[0] = append(records[0], today)

	var wg sync.WaitGroup

	for i, record := range records[1:] {
		wg.Add(1)
		go func(i int, record []string) {
			defer wg.Done()
			name := record[0]
			count, _ := fetchTagCount(name)
			log.Println(name, count)
			output[i+1] = append(record, count)
		}(i, record)
	}

	wg.Wait()

	return output, nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Invalid args: CSV file required")
	}
	filename := os.Args[1]

	input, err := readCsv(filename)
	if err != nil {
		log.Fatal(err)
	}

	output, err := update(input)
	if err != nil {
		log.Fatal(err)
	}

	err = writeCsv(filename, output)
	if err != nil {
		log.Fatal(err)
	}
}
