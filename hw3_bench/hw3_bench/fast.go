package main

import (
	"io"
	"os"
	"io/ioutil"
	"regexp"
	"strings"
	"encoding/json"
	"fmt"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	r := regexp.MustCompile("@")
	seenBrowsers := make(map[string] struct{})
	foundUsers := ""

	lines := strings.Split(string(fileContents), "\n")

	users := make([]map[string]interface{}, 0)
	for _, line := range lines {
		user := make(map[string]interface{})
		// fmt.Printf("%v %v\n", err, line)
		err := json.Unmarshal([]byte(line), &user)
		if err != nil {
			panic(err)
		}
		fmt.Println(user["browsers"])
		users = append(users, user)
	}

	for i, user := range users {

		isAndroid := false
		isMSIE := false

		browsers, ok := user["browsers"].([]interface{})
		if !ok {
			// log.Println("cant cast browsers")
			continue
		}

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			if !ok {
				// log.Println("cant cast browser to string")
				continue
			}

			android := strings.Contains(browser, "Android")
			msie := strings.Contains(browser, "MSIE")

			if android || msie {
				if android {
					isAndroid = true
				}

				if msie {
					isMSIE = true
				}

				if _, ok := seenBrowsers[browser]; !ok {
					seenBrowsers[browser] = struct{}{}
				}
			}
		}

		//for _, browserRaw := range browsers {
		//	browser, ok := browserRaw.(string)
		//	if !ok {
		//		// log.Println("cant cast browser to string")
		//		continue
		//	}
		//	if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
		//		isMSIE = true
		//		notSeenBefore := true
		//		for _, item := range seenBrowsers {
		//			if item == browser {
		//				notSeenBefore = false
		//			}
		//		}
		//		if notSeenBefore {
		//			// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
		//			seenBrowsers = append(seenBrowsers, browser)
		//			uniqueBrowsers++
		//		}
		//	}
		//}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := r.ReplaceAllString(user["email"].(string), " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}