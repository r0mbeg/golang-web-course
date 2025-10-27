package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
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

func getInt(q url.Values, key string, def int) int {
	if v := q.Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func SearchServer(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/favicon.ico" {
		http.NotFound(w, r)
		return
	}

	q := r.URL.Query()
	query := q.Get("query")
	orderField := q.Get("order_field")
	orderBy := getInt(q, "order_by", 0)
	limit := getInt(q, "limit", 10) // дефолт 10
	offset := getInt(q, "offset", 0)

	users, err := users(query, orderField, orderBy, limit, offset)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if errors.Is(err, errBadOrderField) {
			_ = json.NewEncoder(w).Encode(SearchErrorResponse{Error: "ErrorBadOrderField"})
		} else {
			_ = json.NewEncoder(w).Encode(SearchErrorResponse{Error: err.Error()})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(users)

}

type lessFunc func(i, j int) bool

var errBadOrderField = errors.New("bad order_field")

func users(query, orderField string, orderBy, limit, offset int) ([]User, error) {

	//fmt.Printf("Called SearchServer with params: query=%s, orderField=%s, orderBy=%d, limit=%d, offset=%d\n", query, orderField, orderBy, limit, offset)

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
					About:  strings.TrimSpace(userRaw.About), // cut \n, \r and spaces in the end
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
		return nil, errBadOrderField
	}

	switch orderBy {
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

	//fmt.Printf("Got %d filtered users\n", len(users[offset:end]))

	return users[offset:end], nil

}

type TestCase struct {
	Request  SearchRequest
	Response SearchResponse
	IsError  bool
}

func TestSearchClient_FindUsers(t *testing.T) {
	cases := []TestCase{
		{
			Request: SearchRequest{
				Query:      "Boyd",
				Limit:      10,
				Offset:     0,
				OrderField: "",
				OrderBy:    0,
			},
			Response: SearchResponse{
				Users: []User{
					User{
						Id:     0,
						Name:   "Boyd Wolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			IsError: false,
		},
		{
			Request: SearchRequest{
				Query:      "Boydd",
				Limit:      10,
				Offset:     0,
				OrderField: "About",
				OrderBy:    0,
			},
			Response: SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	for i, testCase := range cases {
		sc := &SearchClient{
			AccessToken: "token",
			URL:         ts.URL,
		}

		result, err := sc.FindUsers(testCase.Request)

		// unexpected error
		if err != nil {
			if !testCase.IsError {
				t.Errorf("[%d]: Expected no error, got %s", i, err)
			}
			continue
		}

		// no expected error
		if testCase.IsError {
			t.Errorf("[%d]: Expected error, got no error", i)
			continue
		}

		// only if err == nil, check results
		if len(result.Users) != len(testCase.Response.Users) {
			t.Errorf("[%d]: Expected %d users, got %d", i, len(testCase.Response.Users), len(result.Users))
		}
		for i, user := range result.Users {
			if !reflect.DeepEqual(user, testCase.Response.Users[i]) {
				t.Errorf("[%d]: Expected User %v, got %v", i, testCase.Response.Users[i], user)
			}
		}

	}

}
