package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "net/url"
    "regexp"
    "strconv"
    "strings"
    "sort"

    "github.com/urfave/negroni"
)

type Book struct {
    Title string `json:"title"`
    AuthorName []string `json:"author_name"`
    ReleaseYear int `json:"first_publish_year"`
    Languages []string `json:"language"`
    	
}

type OpenLibraryResponse struct {
    Docs []Book `json:"docs"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

    var searchName string
    var year int
    var operator string
    var isRange bool = false

    var startYear int
    var endYear int

    query := r.URL.Query().Get("q")
    if query == "" {
        log.Println("Missing 'q' parameter")
        http.Error(w, "Missing query parameter `q`", http.StatusBadRequest)
        return
    }

    pageStr := r.URL.Query().Get("page")
    pageSizeStr := r.URL.Query().Get("pageSize")
    paginate := false
    page := 1
    pageSize := 10

    if pageStr != "" {
        if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
            page = p
            paginate = true
        }
    }

    if pageSizeStr != "" {
        if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
            pageSize = ps
        }
    }

    reRange := regexp.MustCompile(`(?i)(.*?)\s+year\s*:\s*(\d{4})-(\d{4})`)
    reSingle := regexp.MustCompile(`(?i)(.*?)\s+year\s*([<>=])\s*(\d{4})`)

    sortDec := strings.Contains(strings.ToLower(query), "sort:dec")
    sortAlpha := strings.Contains(strings.ToLower(query), "sort:alph")

    matches := reRange.FindStringSubmatch(query)
    if len(matches) == 4 {
        searchName = strings.Trim(matches[1], "\" ")
        startYear, _ = strconv.Atoi(matches[2])
        endYear, _ = strconv.Atoi(matches[3])
        isRange = true
    } else {
        matches = reSingle.FindStringSubmatch(query)
        if len(matches) == 4 {
            searchName = strings.Trim(matches[1], "\" ")
            operator = matches[2]
            year, _ = strconv.Atoi(matches[3])
            log.Println("Single filter - term:", searchName, "op:", operator, "year:", year)
        } else {
            searchName = strings.Trim(query, "\" ")
        }
    }

    q := url.QueryEscape(searchName)
    url := fmt.Sprintf("http://openlibrary.org/search.json?q=%s", q)
    log.Println("Searching:", url)

    resp, err := http.Get(url)
    if err != nil {
        log.Println("GET Error:", err)
        http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("Error reading response:", err)
        http.Error(w, "Failed to read response", http.StatusInternalServerError)
        return
    }

    var libResp OpenLibraryResponse
    if err := json.Unmarshal(body, &libResp); err != nil {
        log.Println("JSON parse Error:", err)
        http.Error(w, "Failed to parse JSON", http.StatusInternalServerError)
        return
    }

    results := []Book{}
    for _, book := range libResp.Docs {
        log.Println("book:", book)
        //Sometimes the API returns titles that don't contain the correct name
        //this ensures that all returned titles have the requested name
        if strings.Contains(strings.ToLower(book.Title), strings.ToLower(searchName)) {

            if isRange {
                if book.ReleaseYear < startYear || book.ReleaseYear > endYear {
                    continue
                }
            } else if operator != "" {
                switch operator {
                case "<":
                    if !(book.ReleaseYear < year) {
                        continue
                    }
                case ">":
                    if !(book.ReleaseYear > year) {
                        continue
                    }
                case "=":
                    if !(book.ReleaseYear == year) {
                        continue
                    }
                }
            }
            results = append(results, book)
        }
    }

    if sortAlpha {
        sort.Slice(results, func(i, j int) bool {
            return strings.ToLower(results[i].Title) < strings.ToLower(results[j].Title)
        })
    } else{
        sort.Slice(results, func(i, j int) bool {
            if sortDec {
                return results[i].ReleaseYear > results[j].ReleaseYear
            }
            return results[i].ReleaseYear < results[j].ReleaseYear
        })
    }

    total := len(results)
    if paginate {

        start := (page - 1) * pageSize
        end := start + pageSize

        if start >= total {
            // Invalid page, return response to access right message
            noBooks := Book{
                Title:       "[No books on this page]",
                AuthorName:  []string{"N/A"},
                ReleaseYear: 0,
                Languages:   []string{"N/A"},
            }

            response := map[string]interface{}{
                "total": total,
                "books": []Book{noBooks},
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(response)
            return
        } else {
            if end > total {
                end = total
            }
            results = results[start:end]
        }
    }

    response := map[string]interface{}{
        "total": total,
        "books": results,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)

}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/search", searchHandler)

    n := negroni.Classic()
    n.UseHandler(mux)

    log.Println("Server running on port 3000")
    http.ListenAndServe(":3000", n)
}
