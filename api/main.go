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

    query := r.URL.Query().Get("q")
    if query == "" {
        log.Println("Missing 'q' parameter")
        http.Error(w, "Missing query parameter `q`", http.StatusBadRequest)
        return
    }

    //If user wants to filter by year
    re := regexp.MustCompile(`(?i)(.*?)\s+year\s*([<>])\s*(\d{4})`)
    matches := re.FindStringSubmatch(query)

    if len(matches) == 4 {
        searchName = strings.Trim(matches[1], "\" ")
        operator = matches[2]
        year, _ = strconv.Atoi(matches[3])
        log.Println("Filtro detectado - termo:", searchName, "op:", operator, "ano:", year)
    } else {
        searchName = strings.Trim(query, "\" ")
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

    //log.Println("Resp:", string(body))

    var libResp OpenLibraryResponse
    if err := json.Unmarshal(body, &libResp); err != nil {
        log.Println("JSON parse Error:", err)
        http.Error(w, "Failed to parse JSON", http.StatusInternalServerError)
        return
    }

    results := []Book{}
    for _, book := range libResp.Docs {
        
        //Sometimes the API returns titles that don't contain the correct name
        //this ensures that all returned titles have the requested name
        if strings.Contains(strings.ToLower(book.Title), strings.ToLower(searchName)) {
            if operator != "" {
                if operator == "<" && !(book.ReleaseYear < year) {
                    continue
                }
                if operator == ">" && !(book.ReleaseYear > year) {
                    continue
                }
            }
            results = append(results, book)
        }
    }

    //log.Println("results", results)

    json.NewEncoder(w).Encode(results)
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/search", searchHandler)

    n := negroni.Classic()
    n.UseHandler(mux)

    log.Println("Server running on port 3000")
    http.ListenAndServe(":3000", n)
}
