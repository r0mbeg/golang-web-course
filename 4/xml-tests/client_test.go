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
	"time"
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
				OrderBy:    OrderByAsIs,
			},
			Response: SearchResponse{
				Users: []User{
					{
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
				Query:      "et ea",
				Limit:      10,
				Offset:     0,
				OrderField: "",
				OrderBy:    OrderByDesc,
			},
			Response: SearchResponse{
				Users: []User{
					{
						Id:     33,
						Name:   "Twila Snow",
						Age:    36,
						About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.",
						Gender: "female",
					},
					{
						Id:     7,
						Name:   "Leann Travis",
						Age:    34,
						About:  "Lorem magna dolore et velit ut officia. Cupidatat deserunt elit mollit amet nulla voluptate sit. Quis aute aliquip officia deserunt sint sint nisi. Laboris sit et ea dolore consequat laboris non. Consequat do enim excepteur qui mollit consectetur eiusmod laborum ut duis mollit dolor est. Excepteur amet duis enim laborum aliqua nulla ea minim.",
						Gender: "female",
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
				OrderBy:    OrderByAsIs,
			},
			Response: SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: true,
		},
		{
			Request: SearchRequest{
				Query:      "Boydd",
				Limit:      -1,
				Offset:     0,
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
			Response: SearchResponse{},
			IsError:  true,
		},
		{
			Request: SearchRequest{
				Query:      "Boydd",
				Limit:      26,
				Offset:     0,
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
			Response: SearchResponse{},
			IsError:  false,
		},
		{
			Request: SearchRequest{
				Query:      "Boydd",
				Limit:      10,
				Offset:     -10,
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
			Response: SearchResponse{},
			IsError:  true,
		},
		{
			Request: SearchRequest{
				Query:      "", // можно оставить как есть
				Limit:      1,  // будет обрезан до 25
				Offset:     0,
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
			Response: SearchResponse{
				Users: []User{
					{
						Id:     0,
						Name:   "Boyd Wolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
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

func timeoutErrorHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
}

func unauthorizedErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
}

func internalErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func badRequestErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("`Bad Request}`"))
}

func unknownBadRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(SearchErrorResponse{Error: "SomethingElse"})
}

func cantUnpackJSONErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("`Not encodable json}`")
}

func TestSearchClient_FindUsers_Errors(t *testing.T) {

	t.Run("timeout", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(timeoutErrorHandler))
		defer ts.Close()

		sc := &SearchClient{
			URL:         ts.URL,
			AccessToken: "token",
		}
		_, err := sc.FindUsers(SearchRequest{})
		if err == nil || !strings.Contains(err.Error(), "timeout for") {
			t.Errorf("expected timeout error, got %v", err)
		}
	})

	t.Run("unknown", func(t *testing.T) {

		sc := &SearchClient{
			URL:         "http://localhost:8888888",
			AccessToken: "token",
		}
		_, err := sc.FindUsers(SearchRequest{})
		if err == nil || !strings.Contains(err.Error(), "unknown error") {
			t.Errorf("expected unknown error, got %v", err)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(unauthorizedErrorHandler))
		defer ts.Close()

		sc := &SearchClient{
			URL:         ts.URL,
			AccessToken: "token",
		}

		_, err := sc.FindUsers(SearchRequest{})
		if err == nil || !strings.Contains(err.Error(), "Bad AccessToken") {
			t.Errorf("expected Bad AccessToken, got %v", err)
		}

	})

	t.Run("internal_server_error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(internalErrorHandler))
		defer ts.Close()

		sc := &SearchClient{
			URL:         ts.URL,
			AccessToken: "token",
		}

		_, err := sc.FindUsers(SearchRequest{})
		if err == nil || !strings.Contains(err.Error(), "SearchServer fatal error") {
			t.Errorf("expected SearchServer fatal error, got %v", err)
		}
	})

	t.Run("cant unpack error json", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(badRequestErrorHandler))
		defer ts.Close()

		sc := &SearchClient{
			URL:         ts.URL,
			AccessToken: "token",
		}

		_, err := sc.FindUsers(SearchRequest{})
		if err == nil || !strings.Contains(err.Error(), "cant unpack error json") {
			t.Errorf("expected cant unpack error json, got %v", err)
		}
	})

	t.Run("unknown bad request error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(unknownBadRequestHandler))
		defer ts.Close()

		sc := &SearchClient{
			URL:         ts.URL,
			AccessToken: "token",
		}

		_, err := sc.FindUsers(SearchRequest{})
		if err == nil || !strings.Contains(err.Error(), "unknown bad request error") {
			t.Errorf("expected unknown bad request error, got %v", err)
		}
	})

	t.Run("cant unpack result json", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(cantUnpackJSONErrorHandler))
		defer ts.Close()

		sc := &SearchClient{
			URL:         ts.URL,
			AccessToken: "token",
		}

		_, err := sc.FindUsers(SearchRequest{})
		if err == nil || !strings.Contains(err.Error(), "cant unpack result json") {
			t.Errorf("expected cant unpack result json, got %v", err)
		}

	})

}
