# runner/main.py

import os
import sys
import re
import requests
import math

# Function to print help comand
def print_help():
    print('For Search books use: \'"book name"\'')
    print('For Search books with filter by year [<,>] use: \'"book name" year>yyyy\'')
    print('For Search books in a interval of years [:] use: \'"book name" year:yyyy-yyyy\'')
    print('For Search books and order the list by: crescent (publish year, default), decrescent (publish year), alphabetical (book name) [dec, alph] use: \'"book name" dec\'')
    print('For Search books and separate them into pages (default 10 books per page) [page] use: \'"book name" page=2\'')
    print('For Search books, separate them into pages and change the number of books per page [pageSize] use: \'"book name" page=2 pageSize=20\'')
    print('For Search books and limit the number of books returned [limit] use: \'"book name" limit=5\'')

# Function to print book list
def print_book_list(bookList, start_index=1):
    for i, info in enumerate(bookList, start=start_index):
        title = info.get("title")

        # There are cases where authors, years and languages are empty
        authors_list = info.get("author_name")
        if isinstance(authors_list, list) and len(authors_list) > 0:
            authors = ", ".join(authors_list)
        else:
            authors = "Unknown Author"

        year = info.get("first_publish_year")
        if isinstance(year, int) and year > 0:
            publiYear = str(year)
        else:
            publiYear = "Unknown Year"

        languages = info.get("language")
        if isinstance(languages, list) and len(languages) > 0:
            lang = ", ".join(languages)
        else:
            lang = "Unknown languages"

        print(f"{i}. {title}\n    - authors: {authors}\n    - publish year: {publiYear}\n    - available languages: {lang}\n")

# Clear terminal screen when running the runner
os.system("cls" if os.name == "nt" else "clear")

# Handling user input errors
if len(sys.argv) < 2:
    print('Error: Invalid entry, please follow the examples below:')
    print_help()
    sys.exit(1)

# Print a list of comands if -h or --help is sent
if sys.argv[1] in ("--help", "-h"):
    print_help()
    sys.exit(0)

# Get user inputs
args = sys.argv[1:]
userInput = " ".join(args)

# Get book name
titleMatch = re.search(r'"([^"]+)"', userInput)

if titleMatch:
    bookName = titleMatch.group(1)
else:
    print('Error: Invalid entry, please follow the examples below:')
    print_help()
    sys.exit(1)

# Verify and set year filter if user want to use
operator=""
yearFilter=""

rangeMatch = re.search(r'year\s*:\s*(\d+)-(\d+)', userInput, re.IGNORECASE)
singleMatch = re.search(r'year\s*([<>=])\s*(-?\d+)', userInput, re.IGNORECASE)

if rangeMatch:
    startYear = int(rangeMatch.group(1))
    endYear = int(rangeMatch.group(2))
    if startYear < 1000 or endYear < 1000:
        print("Error: Years in range must be 4-digit positive numbers.")
        sys.exit(1)

    if startYear > endYear:
        startYear = rangeMatch.group(2)
        endYear = rangeMatch.group(1)
        
    yearFilter = f'Range {startYear}-{endYear}'

elif singleMatch:
    operator = singleMatch.group(1)
    year = int(singleMatch.group(2))
    if year < 1000:
        print("Error: Year must be a 4-digit positive number.")
        sys.exit(1)
    yearFilter = f"single {operator} {year}"


page = 1
paginated = False
pageSize = 10

# Verify and set pagination if user want to use
pageMatch = re.search(r'page\s*=\s*(-?\d+)', userInput, re.IGNORECASE)
if pageMatch:
    page = int(pageMatch.group(1))
    if page <= 0:
        print("Error: Page number must be a positive integer.")
        sys.exit(1)
    paginated = True

# If user want to use pagination, verify page size
pageSizeMatch = re.search(r'pagesize\s*=\s*(-?\d+)', userInput, re.IGNORECASE)
if pageSizeMatch:
    pageSize = int(pageSizeMatch.group(1))
    if pageSize <= 0:
        print("Error: Page size must be a positive integer.")
        sys.exit(1)

# Verify and set limit if user want to use
limitMatch = re.search(r'limit\s*=\s*(\d+)', userInput, re.IGNORECASE)
limit = None
if limitMatch:
    limit = int(limitMatch.group(1))

# Verify and set order if user want to use
order=""
decrescent = "dec" in userInput.lower()
alphaOrder = "alph" in userInput.lower()

if decrescent:
    order = "dec"

if alphaOrder:
    order = "alph"

# Set request url to go api
if paginated:
    url = f"http://book-api:3000/search?q={bookName}&year={yearFilter}&order={order}&page={page}&pageSize={pageSize}&limit={limit}"
else:
    url = f"http://book-api:3000/search?q={bookName}&year={yearFilter}&order={order}&limit={limit}"

# Set message to correspondent filter
textFilter=""
if rangeMatch:
    textFilter = f" - Published between {startYear} and {endYear}"
elif singleMatch:
    match operator:
        case "<":
            textFilter=" - Published before " + str(year)
        case ">":
            textFilter=" - Published after " + str(year)
        case "=":
            textFilter=" - Published in " + str(year)
        case _:
            textFilter=""

try:
    # Get go api response and set total and books
    response = requests.get(url)
    response.raise_for_status()
    result = response.json()

    total = result.get("total", 0)
    bookList = result["books"]
       
    # Error message if no book found
    if not bookList:
        print(f"No books found with the name {bookName}")
    else:
        
        # If user want pagination
        if paginated:
            totalPages = math.ceil(total / pageSize)

            # Message if the user wants to access a page that doesn't exist
            if page > totalPages:
                print(f"Page {page} does not exist. There are only {totalPages} pages available.\n")
                sys.exit(1)

            print(f"{total} Books found for: {bookName} {textFilter}\n")

            if totalPages == 1:
                print(f"Showing page 1 of 1 — only one page of results.\n")
            else:
                if page < totalPages:
                    print(f"Page {page} of {totalPages} — use 'page={page+1}' for next page.\n")
                else:
                    print(f"Showing page {page} of {totalPages} — use 'page={page-1}' for previous page.\n")
            
            startIndex = (page - 1) * pageSize + 1
            endIndex = min(startIndex + len(bookList) - 1, total)

            print(f"Showing books {startIndex}–{endIndex} of {total}.\n")
            print_book_list(bookList, startIndex)     

        else:
            print(f"\n{len(bookList)} Books found for: {bookName}{textFilter}\n")
            print_book_list(bookList)

# Error handling
except requests.RequestException as e:
    print("Something went wrong, the requested book could not be found.")
