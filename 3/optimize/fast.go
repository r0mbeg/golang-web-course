package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	androidPattern = "Android"
	iePattern      = "MSIE"
)

type User struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Browsers []string `json:"browsers"`
}

func FastSearch(out io.Writer) {

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	//sc.Buffer(make([]byte, 1024), 1*1024*1024)

	browsers := make(map[string]struct{})
	usersCount := -1
	var foundUsers strings.Builder
	foundUsers.Grow(1 << 16)
	foundUsers.WriteString("found users:\n")

	var u User

	for {
		u.Browsers = u.Browsers[:0]

		if err := dec.Decode(&u); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		usersCount++

		isAndroid, isMSIE := false, false
		for _, b := range u.Browsers {
			a := strings.Contains(b, androidPattern)
			m := strings.Contains(b, iePattern)
			if a || m {
				browsers[b] = struct{}{}
			}
			isAndroid = isAndroid || a
			isMSIE = isMSIE || m
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email := strings.ReplaceAll(u.Email, "@", " [at] ")
		foundUsers.WriteString("[" + strconv.Itoa(usersCount) + "] " + u.Name + " <" + email + ">" + "\n")
		//foundUsers.WriteString(fmt.Sprintf("[%d] %s <%s>\n", usersCount, u.Name, email))
	}

	fmt.Fprintln(out, foundUsers.String())
	fmt.Fprintln(out, "Total unique browsers", len(browsers))

}
