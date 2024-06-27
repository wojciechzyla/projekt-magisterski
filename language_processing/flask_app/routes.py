from flask import Flask, request, render_template, redirect, url_for
from datetime import datetime, timedelta
from pymongo import MongoClient
import os
from qdrant_client import QdrantClient
from langchain_openai import OpenAIEmbeddings, ChatOpenAI
from categories_extractors import (CategoriesExtraction, TermsExtractor, SimilarityAndTermsExtractorWrapper,
                                   SimilarityExtractor)
import requests
import json

app = Flask(__name__)
app.secret_key = 'your_secret_key'

MONGO_URL = os.getenv("MONGO_URL")
QDRANT_URL = os.getenv("QDRANT_URL")
ORIENTEERING_URL = os.getenv("ORIENTEERING_URL")

mongo_client = MongoClient(MONGO_URL)
db = mongo_client.travel_agency
poi_collection = db.pois
t2c_collection = db.term_to_category_map

qdrant = QdrantClient(url=QDRANT_URL)

embeddings_model = OpenAIEmbeddings(model="text-embedding-ada-002")
llm = ChatOpenAI(model_name="gpt-3.5-turbo-0125", temperature=0.7)

@app.route('/', methods=['GET', 'POST'])
def travel_plan():
    if request.method == 'GET':
        return render_template('plan_form.html')
    else:
        json_data = request.form.to_dict()

        start_date = datetime.strptime(json_data.get('startDate'), '%Y-%m-%d')
        end_date = datetime.strptime(json_data.get('endDate'), '%Y-%m-%d')
        day_count = (end_date - start_date).days + 1
        days = [(start_date + timedelta(days=i)).strftime('%a').lower() for i in range(day_count)]
        formatted_days = [(start_date + timedelta(days=i)).strftime('%d-%m - %a') for i in range(day_count)]

        terms_to_categories = list(t2c_collection.find())[0]
        terms_extractor = TermsExtractor(terms_to_categories)
        similarity_extractor = SimilarityExtractor(embeddings_model, qdrant, "categories", llm)
        sim_term_extractor = SimilarityAndTermsExtractorWrapper(similarity_extractor=similarity_extractor,
                                                                terms_extractor=terms_extractor)
        categories_extractor = CategoriesExtraction([sim_term_extractor])

        extracted_categories = categories_extractor.extract_categories(json_data["description"])
        all_pois = list(poi_collection.find({}))

        pois_pool = []
        categories_list = list(extracted_categories.keys())
        used_poi = []
        for poi in all_pois:
            for cat in poi["categories"]:
                if cat in categories_list and poi["name"] not in used_poi:
                    new_poi = {
                        "name": poi["name"],
                        "lon": poi["lon"],
                        "lat": poi["lat"],
                        "satisfaction": extracted_categories[cat],
                        "openHour": poi["openHour"],
                        "closeHour": poi["closeHour"]
                    }
                    pois_pool.append(new_poi)
                    used_poi.append(poi["name"])

        headers = {"Content-Type": "application/json"}
        post_json_data = {"poiList": pois_pool, "days": days,
                          "dayStart": json_data["dayStartHour"], "dayEnd": json_data["dayEndHour"]}

        print("waiting for response")
        response = requests.post(ORIENTEERING_URL, headers=headers, json=post_json_data)
        data = json.loads(json.dumps(response.json()))
        for i in range(len(data["Days"])):
            data["Days"][i]["date"] = formatted_days[i]
        return render_template('itinerary.html', days=data["Days"])
        # return redirect(url_for('display_itinerary', data=json.dumps(response.json()),
        #                         days=json.dumps(formatted_days)))

@app.route('/itinerary', methods=['GET'])
def display_itinerary():
    data = json.loads(request.args.get("data", ""))
    days = json.loads(request.args.get("days", ""))
    for i in range(len(data["Days"])):
        data["Days"][i]["date"] = days[i]
    return render_template('itinerary.html', days=data["Days"])


if __name__ == '__main__':
    app.run(debug=False, port=5000)
