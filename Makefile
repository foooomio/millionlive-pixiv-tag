.PHONY: update
update:
	go run main.go millionlive.csv

.PHONY: clean
clean:
	git checkout -- millionlive.csv
