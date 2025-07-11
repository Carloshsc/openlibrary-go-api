

# runner/main.py

import os
import sys
import requests

if len(sys.argv) < 2:
    print("Usage: runner <search terms>")
    sys.exit(1)

query = sys.argv[1]
port = os.environ.get("API_PORT", "3000")
host = os.environ.get("API_HOST", "api")
url = f"http://{host}:{port}/search?q={query}"
print(f"url' {url}':")

try:
    response = requests.get(url)
    response.raise_for_status()
    bookList = response.json()

    print(f"Books found for '{query}':\n")
    
    if not bookList:
        print(f"No books found with that name")
    else:
        for i, info in enumerate(bookList, 1):
            title = info.get("title", "Unknown Title")

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
    print("Error querying the API:", e)
