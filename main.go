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

const baseURL = "https://dic.pixiv.net/a/"
const maxWorkers = 5

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
	data, err := fetch(baseURL + name)
	if err != nil {
		return "", err
	}

	match := re.FindSubmatch(data)
	if match == nil {
		return "", fmt.Errorf("not match: %s", name)
	}
	return string(match[1]), nil
}

func update(records [][]string) error {
	today := time.Now().Format("2006-01-02")
	records[0] = append(records[0], today)

	limit := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for i, record := range records[1:] {
		limit <- struct{}{}
		wg.Add(1)

		go func(i int, record []string) {
			defer wg.Done()

			name := record[0]
			count, err := fetchTagCount(name)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(name, count)
			records[i+1] = append(record, count)

			<-limit
		}(i, record)
	}

	wg.Wait()

	return nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Invalid args: CSV file required")
	}
	filename := os.Args[1]

	records, err := readCsv(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = update(records)
	if err != nil {
		log.Fatal(err)
	}

	err = writeCsv(filename, records)
	if err != nil {
		log.Fatal(err)
	}
}
