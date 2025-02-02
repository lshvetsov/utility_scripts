# Description

The script that extends the list of movies with data from Kinopoisk API.

# Usage

## Arguments 

1. *Key* - API key to fetch data from Kinopoisk, required. 
    It should be placed in the `key.txt` file in the current directory or provided as an argument (if running the script from python).
2. *List of movies* - file with list of movies, optional.   
    The default source file path is `movies.txt` in the current directory but it can be changed using the `--list` argument (or with script questions in case of bat and sh scripts).
3. Output file - file to write the extended list of movies, optional.   
    The default target file is `movies.csv` in the current directory but it can be changed using the `--output` argument (or with script questions in case of bat and sh scripts).

## Examples

- `movies.bat` - Windows batch script
- `movies.sh` - Unix shell script
- Python script:

```shell
movie_processor.py --key <key> --list <list of movies> --output <output file>
```


# Input data format

File with list of movies in the following format:
```
Movie name (Original name, Year)
Movie name (Original name, Year)
```

# Output data format

CSV file with extended list of movies (for uploading to Notion).
