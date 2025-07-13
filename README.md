![](https://img.shields.io/badge/Status-Working-brightgreen) 
![](https://img.shields.io/badge/Release-1.0-blue)
# Open Library Go API

This project is a book search service powered by the OpenLibrary API. It returns a list of books containing the title, author(s), publication year, and available languages. The backend is built in Go and communicates with the external API, while a Python-based runner handles user interaction. Both components are fully containerized using Docker.


## Instructions

This project uses Docker for containerization, so it’s necessary to build the required images before running anything.

To build the images, run the following commands:

First, go to the `openlibrary-go-api/api` folder and run:
`docker compose build api`

Then, go to the `openlibrary-go-api/runner` folder and run:
`docker compose build runner`

Once the images are built, you can run the application using the following steps, in this specific order:

First, start the API container by running: `docker compose up api`

Then run the Runner container with the desired book: `docker compose run --rm runner '"The Epic of Gilgamesh"'`

Make sure that the API container is up before executing the runner, as the runner depends on it to retrieve data from the OpenLibrary API and format the output.

It's important to note that when running the runner command, the entire input must be enclosed in single quotes ('), and the book title itself must be inside double quotes ("). This formatting is required to correctly handle special characters such as <, >, and : without causing issues in the terminal.

Below are the supported command options you can use alongside the book title to filter or customize the output:

-By default, books are sorted in crescenting publication year order.

| Option           | Description                                   |
| ---------------- | --------------------------------------------- |
| `"Book Title"`   | Title to search (must be in double quotes)    |
| `year>YYYY`      | Filter: published after the year              |
| `year<YYYY`      | Filter: published before the year             |
| `year=YYYY`      | Filter: published in the year                 |
| `year:YYYY-YYYY` | Filter: published within year range           |
| `dec`            | Sort by publication year decreasing           |
| `alph`           | Sort by title (alphabetical order)            |
| `page=X`         | Select page number (default: 1)               |
| `pageSize=Y`     | Results per page (default: 10)                |
| `limit=Z`        | Limit total number of results                 |
