package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "sort"
    "github.com/urfave/negroni"
)

// Book structure with necessary information
type Book struct {
    Title string `json:"title"`
    AuthorName []string `json:"author_name"`
    ReleaseYear int `json:"first_publish_year"`
    Languages []string `json:"language"`
}

// An array that will store books
type OpenLibraryResponse struct {
    Docs []Book `json:"docs"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {    

    // Get the info passed by the url and set book name
    query := r.URL.Query().Get("q")
    if query == "" {
        log.Println("Missing 'q' parameter")
        http.Error(w, "Missing query parameter `q`", http.StatusBadRequest)
        return
    }

    searchName := strings.TrimSpace(strings.ToLower(query))

    // Get the info passed by the url and set pagination and page size
    // If pagination is set but page size is not, default value will be 10
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
    
    // Get the info passed by the url and set order
    sortParam := r.URL.Query().Get("order")
    sortAlpha := sortParam == "alph"
    sortDec := sortParam == "dec"

    // Get the info passed by the url and set year filter
    yearParam := r.URL.Query().Get("year")

    isRange := false

    var year int
    var startYear int
    var endYear int
    var operator string

    //If is range will use filter between 2 years
    if strings.HasPrefix(yearParam, "Range") {
        parts := strings.Fields(strings.TrimPrefix(yearParam, "Range"))
        if len(parts) == 1 && strings.Contains(parts[0], "-") {
            years := strings.Split(parts[0], "-")
            if len(years) == 2 {
                startYear, _ = strconv.Atoi(years[0])
                endYear, _ = strconv.Atoi(years[1])
                // Adjust year order if user passes largest value first
                if startYear > endYear {
                    startYear = endYear
                    endYear =  startYear
                }
                isRange = true
            }
        }
    } else if strings.HasPrefix(yearParam, "single") {
        // If is a single year set operator
        parts := strings.Fields(strings.TrimPrefix(yearParam, "single"))
        if len(parts) == 2 {
            operator = parts[0]
            year, _ = strconv.Atoi(parts[1])
        }
    }

    // Get the info passed by the url and set limit
    limitStr := r.URL.Query().Get("limit")   

    // Create the url to search the book
    q := url.QueryEscape(searchName)
    url := fmt.Sprintf("http://openlibrary.org/search.json?q=%s", q)
    log.Println("Searching:", url)

    // Get the response
    resp, err := http.Get(url)
    if err != nil {
        log.Println("GET Error:", err)
        http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    // Reads the response
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("Error reading response:", err)
        http.Error(w, "Failed to read response", http.StatusInternalServerError)
        return
    }

    // Store books data into an array
    var libResp OpenLibraryResponse
    if err := json.Unmarshal(body, &libResp); err != nil {
        log.Println("JSON parse Error:", err)
        http.Error(w, "Failed to parse JSON", http.StatusInternalServerError)
        return
    }

    results := []Book{}
    // Get the right books
    for _, book := range libResp.Docs {
        // Sometimes the API returns titles that don't contain the correct name
        // This ensures that all returned titles have the requested name
        if strings.Contains(strings.ToLower(book.Title), strings.ToLower(searchName)) {
            // Filter books by year filter, if set
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

    // Order the books
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

    // If limit is set get the first number of books requested
    total := len(results)
    if limitStr != "" {
        if lim, err := strconv.Atoi(limitStr); err == nil && lim > 0 && lim < total {
            results = results[:lim]
            total = len(results)
        }
    }

    // Create an array for pagination
    if paginate && total > 0{

        start := (page - 1) * pageSize
        end := start + pageSize

        if start >= total {
            // Invalid page, return response to show right message
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

    // Create the response
    response := map[string]interface{}{
        "total": total,
        "books": results,
    }

    // Return the information
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)

}

func main() {

    // Mux start
    mux := http.NewServeMux()

    // Route declaration for /search
    mux.HandleFunc("/search", searchHandler)

    // Negroni default declarations
    n := negroni.Classic()
    n.UseHandler(mux)
    
    // API message to show port that will be used
    log.Println("Server running on port 3000")
    http.ListenAndServe(":3000", n)
}
