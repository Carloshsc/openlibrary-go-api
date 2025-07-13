# runner/main.py

import os
import sys
import re
import requests
import math

if len(sys.argv) < 2:
    print("Usage: runner <search terms>")
    sys.exit(1)

args = sys.argv[1:]

page = 1
paginated = False

userInput = " ".join(args)
pageSize = 10

print(f"Received input: {userInput}")
titleMatch = re.search(r'"([^"]+)"', userInput)

if titleMatch:
    bookName = titleMatch.group(1)
    print(f"titleMatch: {bookName}")
else:
    print("Error: Book title must be inside double quotes (e.g., \"Lord of the Rings\")")
    sys.exit(1)

decrescent = "dec" in userInput.lower()
alphaOrder = "alph" in userInput.lower()

rangeMatch = re.search(r'year\s*:\s*(\d{4})-(\d{4})', userInput, re.IGNORECASE)
singleMatch = re.search(r'year\s*([<>=])\s*(\d{4})', userInput, re.IGNORECASE)

pageMatch = re.search(r'page\s*=\s*(\d+)', userInput, re.IGNORECASE)
if pageMatch:
    page = int(pageMatch.group(1))
    paginated = True

pageSizeMatch = re.search(r'pagesize\s*=\s*(\d+)', userInput, re.IGNORECASE)
if pageSizeMatch:
    pageSize = int(pageSizeMatch.group(1))

operator=""
textFilter=""

if rangeMatch:
    startYear = rangeMatch.group(1)
    endYear = rangeMatch.group(2)
    if startYear > endYear:
        startYear = rangeMatch.group(2)
        endYear = rangeMatch.group(1)
    query = f'{bookName} year:{startYear}-{endYear}'
elif singleMatch:
    operator = singleMatch.group(1)
    year = singleMatch.group(2)
    query = f"{bookName} year {operator} {year}"
else:
    query = bookName

if decrescent:
    query += " sort:dec"

if alphaOrder:
    query += " sort:alph"

port = os.environ.get("API_PORT", "3000")
host = os.environ.get("API_HOST", "api")

if paginated:
    url = f"http://{host}:{port}/search?q={query}&page={page}&pageSize={pageSize}"
else:
    url = f"http://{host}:{port}/search?q={query}"

print(f"url' {url}':")

if rangeMatch:
    textFilter = f" - Published between {startYear} and {endYear}"
elif singleMatch:
    match operator:
        case "<":
            textFilter=" - Published before " + year
        case ">":
            textFilter=" - Published after " + year
        case "=":
            textFilter=" - Published in " + year
        case _:
            textFilter=""

try:
    response = requests.get(url)
    response.raise_for_status()
    result = response.json()

    total = result.get("total", 0)
    bookList = result["books"]
       
    if not bookList:
        print(f"No books found with that name")
    else:
        
        if paginated:
            totalPages = math.ceil(total / pageSize)

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

            for i, info in enumerate(bookList, start=startIndex):
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

        else:
            print(f"\n{len(bookList)} Books found for: {bookName}{textFilter}\n")

            for i, info in enumerate(bookList, 1):
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
except requests.RequestException as e:
    print("Something went wrong, the requested book could not be found.")
