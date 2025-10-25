package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type UserRaw struct {
	Id        int    `xml:"id,attr"`
	FirstName string `xml:"first_name,attr"`
	LastName  string `xml:"last_name,attr"`
	Age       int    `xml:"age,attr"`
	About     string `xml:"about,attr"`
	Gender    string `xml:"gender,attr"`
}

const (
	fileName = "dataset.xml"
)

func main() {
	_, _ = SearchServer("", "", 0, 0, 0)
}
func SearchServer(query, orderField string, orderDirection, limit, offset int) ([]User, error) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	dec := xml.NewDecoder(file)

	users := []User{}

	var userRaw UserRaw

	for {
		if err := dec.Decode(&userRaw); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
	}

}

func MustLoadUsers(fileName string) []User {
	const op = "client_test.LoadUsers"

}
