package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

type UserRaw struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

type dataset struct {
	Rows []UserRaw `xml:"row"`
}

const (
	fileName = "dataset.xml"
)

func main() {
	users, err := SearchServer("Bo", "age", 0, 100, 30)

	if err != nil {
		fmt.Println(err)
	}

	for _, user := range users {
		fmt.Printf("%+v\n", user)
	}
}

type lessFunc func(i, j int) bool

func SearchServer(query, orderField string, orderDirection, limit, offset int) ([]User, error) {

	// open file

	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// decode XML -> dataset
	var ds dataset
	if err := xml.NewDecoder(file).Decode(&ds); err != nil {
		panic(err)
	}

	// fill in users
	users := make([]User, 0, len(ds.Rows))

	for _, userRaw := range ds.Rows {

		name := userRaw.FirstName + " " + userRaw.LastName

		if query == "" ||
			strings.Contains(strings.ToLower(name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(userRaw.About), strings.ToLower(query)) {
			users = append(users,
				User{
					Id:     userRaw.Id,
					Name:   name,
					Age:    userRaw.Age,
					About:  userRaw.About,
					Gender: userRaw.Gender,
				})
		}
	}

	// choose less func
	var less lessFunc
	switch strings.ToLower(orderField) {
	case "", "name":
		less = func(i, j int) bool { return users[i].Name < users[j].Name }
	case "age":
		less = func(i, j int) bool { return users[i].Age < users[j].Age }
	case "id":
		less = func(i, j int) bool { return users[i].Id < users[j].Id }
	default:
		return nil, errors.New(ErrorBadOrderField)
	}

	switch orderDirection {
	case OrderByAsc:
		sort.SliceStable(users, less)
	case OrderByDesc:
		sort.SliceStable(users, func(i, j int) bool { return less(j, i) })
	case OrderByAsIs:

	default:

	}

	// offset checking
	if offset < 0 {
		offset = 0
	}

	if limit <= 0 {
		return []User{}, nil
	}

	// limit checking
	if offset >= len(users) {
		return []User{}, nil
	}
	end := offset + limit
	if end > len(users) {
		end = len(users)
	}

	return users[offset:end], nil

}
