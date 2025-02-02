import requests
import re
import pandas as pd
import argparse
import logging
from tqdm import tqdm

extraction_pattern = re.compile(r"\s*[^a-zA-Z0-9]*\(([^)]+)\)")
cleaning_pattern = re.compile(r"\s*\([^)]*\)$")
base_url = 'https://api.kinopoisk.dev/v1.4/movie/search'
logging.basicConfig(level=logging.DEBUG, format='%(asctime)s - %(levelname)s - %(message)s')


# Read user input
def read_script_parameters():
    parser = argparse.ArgumentParser(description="Get data for processing list of movies")
    parser.add_argument('api_key', type=str, help="API key to fetch data from Kinopoisk")
    parser.add_argument('--list', type=str, default="movies.txt",
                        help="Path to the file with list of movies, default: movies.txt")
    parser.add_argument('--output', type=str, default="movies.csv", help="Path to output csv file, default: movies.csv")

    args = parser.parse_args()
    api_key = args.api_key
    movie_list_path = args.list if args.list else "./movies.txt"
    output_path = args.output if args.output else "./movies.csv"
    return api_key, movie_list_path, output_path


# Reading original list of movies
def read_input_list_to_dict(list_path):
    movies_dict = {}
    with open(list_path, 'r', encoding='utf-8') as file:
        for line in file:
            line = line.strip()
            name, details = extract_details(line)
            if name:  # Check that name isn't empty
                movies_dict[name] = details
                logging.debug(f"Read: {name} with {details}")
            else:
                logging.warning(f"Skipped line with empty movie name: {line}")
    logging.debug(f"Total movies in the list: {len(movies_dict)}")
    return movies_dict


def extract_details(line):
    match_extraction = extraction_pattern.search(line)
    details = match_extraction.group(1) if match_extraction else ""
    clean_name = cleaning_pattern.sub("", line).strip()

    if "," in details:
        original, year = details.split(",", 1)
    elif any(char.isdigit() for char in details):
        original = ""
        year = int(details)
    else:
        original = details
        year = 0

    return clean_name, {"Original": original.strip(), "Year": year}


# Fetching and choose the proper Kinopoisk card

def refine_movie_dict(movie_dict, api_key):
    if movie_dict is None:
        raise Exception("No movies found in the list")

    for name, data in tqdm(movie_dict.items(), desc="Processing movies", unit="movie"):
        original = data.get("Original")
        year = int(data.get("Year"))
        movies_data = fetch_movie_data(api_key, name)['docs']
        target_movie = auto_match(movies_data, name, original, year)
        if not target_movie:
            print(f"{name}: no match found, you need to manually opt for the target")
            target_movie = manual_match(name, movies_data)
        movie_dict[name] = transform_data_for_processing(target_movie)


def fetch_movie_data(api_key, movie_name, page=1, limit=3):
    headers = {
        'Accept': 'application/json',
        'X-API-KEY': api_key
    }
    params = {
        'query': movie_name,
        'page': page,
        'limit': limit
    }
    response = requests.get(base_url, headers=headers, params=params, timeout=20)
    if response.status_code == 200:
        return response.json()
    else:
        raise Exception(f"Failed to fetch movie data. Status code: {response.status_code}; Message: {response.text}")


def auto_match(movies_data, name, original, year):
    for movie in movies_data:
        api_names = extract_names(movie)
        if compare_names(name, original, api_names) and year == movie['year']:
            print(f"{name}: match found, 1 movie")
            return movie


def manual_match(name, candidates):
    print(f"Multiple entries found for '{name}':")
    for index, candidate in enumerate(candidates):
        print(f"{index + 1}. {candidate['name']} ({candidate['year']})")
        print(f"   Countries: {candidate['countries']}")
        print(f"   Genres: {candidate['genres']}")
        print(f"   Candidate ratings: {candidate['rating']}")
    index = int(input(f"Select the correct entry (1-{len(candidates)}): ")) - 1
    return candidates[index]


def extract_names(movie):
    names_list = [movie["name"]]
    if "alternativeName" in movie:
        names_list.append(movie["alternativeName"])
    names_list.extend([name_entry["name"] for name_entry in movie["names"] if "name" in name_entry])
    return names_list


def compare_names(list_name, list_original_name, api_names):
    list_name = convert_name(list_name)
    api_names = [convert_name(name) for name in api_names]
    return list_name in api_names or list_original_name in api_names


def convert_name(name):
    return re.sub(r'\s+', ' ', re.sub(r'[^\w\s]', '', name.lower(), flags=re.UNICODE)).strip()


def transform_data_for_processing(data):
    result = {}
    rating_data = data.get('rating', {})
    imdb_rating = rating_data.get('imdb')
    kp_rating = rating_data.get('kp')
    external_ids = data.get('externalId', {})
    imdb_id = external_ids.get('imdb') if external_ids else None
    kp_id = external_ids.get('kpHD') if external_ids else None

    result['name'] = data.get('name') if data.get('name') else data.get('alternativeName')
    result['countries'] = ', '.join([country['name'] for country in data.get('countries', []) if 'name' in country])
    result['year'] = data.get('year')
    result['genres'] = ', '.join([genre['name'] for genre in data.get('genres', []) if 'name' in genre])
    result['rating_imdb'] = imdb_rating
    result['rating_kp'] = kp_rating
    result['imdb_id'] = imdb_id
    result['kp_id'] = kp_id
    return result


# Export data to CSV to upload to DB

def process_data_for_export(movie_dict, output_path):
    df = pd.DataFrame(movie_dict.values())
    df.to_csv(output_path, index=False, encoding='utf-8')
    print("Data has been written to 'movies.csv'")


def main():
    api_key, movie_list_path, output_path = read_script_parameters()
    movie_dict = read_input_list_to_dict(movie_list_path)
    refine_movie_dict(movie_dict, api_key)
    process_data_for_export(movie_dict, output_path)


if __name__ == "__main__":
    main()
