package main

import (
	"bufio"
	"database/sql"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var database = "pubmed"

type ESearchResult struct {
	XMLName  xml.Name `xml:"eSearchResult"`
	Text     string   `xml:",chardata"`
	Count    string   `xml:"Count"`
	RetMax   string   `xml:"RetMax"`
	RetStart string   `xml:"RetStart"`
	IdList   struct {
		Text string   `xml:",chardata"`
		ID   []string `xml:"Id"`
	} `xml:"IdList"`
	TranslationSet   string `xml:"TranslationSet"`
	TranslationStack struct {
		Text    string `xml:",chardata"`
		TermSet struct {
			Text    string `xml:",chardata"`
			Term    string `xml:"Term"`
			Field   string `xml:"Field"`
			Count   string `xml:"Count"`
			Explode string `xml:"Explode"`
		} `xml:"TermSet"`
		OP string `xml:"OP"`
	} `xml:"TranslationStack"`
	QueryTranslation string `xml:"QueryTranslation"`
}

func pickDatabase() {
	fmt.Println("Select Database: (dbSNP(as: snp) or pubmed)")
	buf := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	input, err := buf.ReadBytes('\n')
	if err != nil {
		fmt.Println(err)
	} else {
		database = strings.TrimSuffix(string(input), "\n")
	}
}

func queryAPI(db, query string) string {
	// address := "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&term=science[journal]+AND+breast+cancer+AND+2008[pdat]"
	address := "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=" + db + "&term=" + query
	// fmt.Println("Address given:", address)
	resp, err := http.Get(address)
	if err != nil {
		// handle err
		fmt.Println("queryAPI didn't work")
	}
	// defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))
	return string(body)
}

func getPubmedIDs(body string) {
	byteValue := []byte(body)
	var esearchresult ESearchResult
	xml.Unmarshal(byteValue, &esearchresult)
	for i := 0; i < len(esearchresult.IdList.ID); i++ {
		fmt.Println("pubmed ID: ", esearchresult.IdList.ID[i])
	}
	return
}

// waits for query to be entered, then executes the search
func checkInput() {
	for {
		fmt.Println("Enter Search Term: (type: switch to change database)")
		buf := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		query, err := buf.ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
		} else {
			term := strings.TrimSuffix(string(query), "\n")
			if term == "switch" {
				pickDatabase()
			} else {
				// fmt.Println(term)
				fmt.Println(queryAPI(database, term))
				if term == "pubmed_snp_cited[sb]" {
					getPubmedIDs(queryAPI("pubmed", term))
				}
			}
		}
	}
}

func createSqlite() {
	os.MkdirAll("./data", os.ModePerm)
	os.Create("./data/ncbidata.db")

	db, err := sql.Open("sqlite3", "./data/ncbidata.db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = db.Exec("CREATE TABLE `pubmed` (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `client_id` VARCHAR(64) NULL, `first_name` VARCHAR(255) NOT NULL, `last_name` VARCHAR(255) NOT NULL, `guid` VARCHAR(255) NULL, `dob` DATETIME NULL, `type` VARCHAR(1), 'FOREIGN KEY(rsids)' REFERNCES rsids(id))")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE `rsids` (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `client_id` VARCHAR(64) NULL, `first_name` VARCHAR(255) NOT NULL, `last_name` VARCHAR(255) NOT NULL, `guid` VARCHAR(255) NULL, `dob` DATETIME NULL, `type` VARCHAR(1), 'FOREIGN KEY(pubmed)' REFERNCES pubmed(id))")
	if err != nil {
		log.Fatal(err)
	}

	db.Close()
}

func main() {
	pickDatabase()
	createSqlite()
	checkInput()
	return
}
