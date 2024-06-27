from pymongo import MongoClient
import argparse
import os
import json
from typing import Union
from dotenv import load_dotenv
from nltk.stem import PorterStemmer
load_dotenv()

MONGO_URL = os.getenv("MONGO_URL")


def human_bool(flag: Union[str, bool], default: bool = False) -> bool:
    if flag is None:
        return False
    if isinstance(flag, bool):
        return flag
    if flag.lower() in [
        "true",
        "1",
        "t",
        "y",
        "yes",
    ]:
        return True
    elif flag.lower() in [
        "false",
        "0",
        "f",
        "n",
        "no",
    ]:
        return False
    else:
        return default


def get_pois_from_file(file_path, categories):
    with open(file_path, "r") as f:
        all_pois = json.load(f)
    result = []
    categories_list = list(categories.keys())
    used_poi = []
    for poi in all_pois:
        for cat in poi["categories"]:
            if cat in categories_list and poi["name"] not in used_poi:
                new_poi = {
                    "name": poi["name"],
                    "lon": poi["lon"],
                    "lat": poi["lat"],
                    "satisfaction": categories[cat],
                    "openHour": poi["openHour"],
                    "closeHour": poi["closeHour"]
                }
                result.append(new_poi)
                used_poi.append(poi["name"])
    return result


def term_to_category_mapping(all_categories):
    stemmer = PorterStemmer()
    categories_terms = {}
    for ic, category in enumerate(all_categories):
        parts = category.split(".")
        for ip, part in enumerate(parts):
            multiword = part.split("_")
            for iw, word in enumerate(multiword):
                multiword[iw] = stemmer.stem(word)
            key = " ".join(multiword)
            if key not in categories_terms:
                categories_terms[key] = [category]
            else:
                add_new_cat = True
                for existing in categories_terms[key]:
                    if existing == ".".join(parts[:ip+1]):
                        add_new_cat = False
                if add_new_cat:
                    categories_terms[key].append(category)
    return categories_terms


def collect_categories(categories: dict):
    cats = []
    for key, val in categories.items():
        cats.append(key)
        if val["subcategories"] is not None:
            subcats = collect_categories(val["subcategories"]["categories"])
            for s in subcats:
                cats.append(f"{key}.{s}")
    return cats


def main():
    parser = argparse.ArgumentParser(description="Load mongoDB collections")
    parser.add_argument("--poi_file_path", help="Path to load pois from. If not provided, pois won't "
                                                "be loaded")
    parser.add_argument("--categories_file_path", help="Path to load to a file with categories. If not "
                                                       "provided, categories won't be loaded")
    args = parser.parse_args()

    poi_file_path = args.poi_file_path
    categories_file_path = args.categories_file_path

    mongo_client = MongoClient(MONGO_URL)
    db = mongo_client.travel_agency

    if os.path.isfile(poi_file_path):
        with open(poi_file_path, "r") as f:
            all_pois = json.load(f)
        print("Inserting pois...")
        pois = db.pois
        pois.insert_many(all_pois)
        print("Pois successfully inserted")

    if os.path.isfile(categories_file_path):
        with open(categories_file_path, "r") as f:
            categories = json.load(f)
        all_categories = collect_categories(categories["categories"])

        term_to_category_map = db.term_to_category_map
        t2c = term_to_category_mapping(all_categories)
        print("Inserting terms to category mapping...")
        term_to_category_map.insert_one(t2c)
        print("Terms to category successfully inserted")


if __name__ == "__main__":
    main()


"""
python3 mongo_loader.py --poi_file_path /Users/wzya/Desktop/Wojtek/AGH/sp2/langchain/manual_pois.json --categories_file_path /Users/wzya/Desktop/Wojtek/AGH/sp2/langchain/places.json
"""