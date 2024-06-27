import argparse
import os
from dotenv import load_dotenv
from typing import Union
from qdrant_client import models, QdrantClient
import json
from langchain_openai import OpenAIEmbeddings, ChatOpenAI
from langchain_core.output_parsers import StrOutputParser
from langchain.prompts import ChatPromptTemplate
load_dotenv()

QDRANT_URL = os.getenv("QDRANT_URL")


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


def load_embeddings_from_file(file_name):
    names = []
    embeddings = []
    with open(file_name, "r") as f:
        data = json.loads(f.read())
        for key, val in data.items():
            names.append(key)
            embeddings.append(val)
    embedding_size = len(embeddings[0])
    return names, embeddings, embedding_size


def load_to_qdrant(names, embeddings, embedding_size):
    #qdrant = QdrantClient(":memory:")
    qdrant = QdrantClient(url=QDRANT_URL)
    qdrant.recreate_collection(
        collection_name="categories",
        vectors_config=models.VectorParams(
            size=embedding_size,
            distance=models.Distance.COSINE,
        ),
    )

    qdrant.upsert(
        collection_name="categories",
        points=[
            models.PointStruct(id=i, vector=embeddings[i], payload={"cat": names[i]}) for i in range(len(embeddings))
        ],
    )

    return qdrant


def collect_categories(categories: dict):
    cats = []
    for key, val in categories.items():
        cats.append(key)
        if val["subcategories"] is not None:
            subcats = collect_categories(val["subcategories"]["categories"])
            for s in subcats:
                cats.append(f"{key}.{s}")
    return cats


def create_descriptions(all_categories: list, llm, descriptions_file_path: str):
    template = """
    I will give you a travel point of interest category. This category might consist of several subcategories. Each 
    subcategory is split by a dot. Example category might look like this: accommodation.cheap.hostel. I want you 
    to write short text describing travel expectations and plans from the point of view of traveller that is going 
    to be a perfect match for the provided category. This description should be 3 sentences long.

    ###

    Provided category: {category}
    """
    prompt = ChatPromptTemplate.from_template(template)
    output_parser = StrOutputParser()
    llm_chain = prompt | llm | output_parser
    categories_names = []
    descriptions = []

    for i, category in enumerate(all_categories):
        response = llm_chain.invoke({"category": category})
        categories_names.append(category)
        descriptions.append(response)
        print(f"Creating description for category {i + 1}/{len(all_categories)}")

    if descriptions_file_path:
        with open(descriptions_file_path, "w") as f:
            result = {}
            for cat, desc in zip(categories_names, descriptions):
                result[cat] = desc
            json.dump(result, f)
    return categories_names, descriptions


def main():
    parser = argparse.ArgumentParser(description="Load categories embeddings to database or file")
    parser.add_argument("--destination", default="d", help="Destination where embeddings will be stored. "
                                                           "'d' for database, 'f' for file, 'df' for database and file."
                                                           " Default is 'd'.")
    parser.add_argument("--destination_file_path", help="Path to the destination file, if 'f' was in destination")
    parser.add_argument("--source_type", help="'embeddings' if embeddings already exist in file. "
                                              "'descriptions' if textual descriptions for categories exist and "
                                              "embeddings should be created from them. 'categories' if descriptions must"
                                              " be previously created for categories")
    parser.add_argument("--source_file", help="File to read data from. For source_type=embeddings, "
                                                         "this will be a file with embeddings."
                                              "For source_type=descriptions, "
                                                         "this will be a file with categories descriptions."
                                              "For source_type=categories, "
                                                         "this will be a file with categories.")
    parser.add_argument("--descriptions_file_path", help="If source_type=categories and this path is "
                                                         "provided, descriptions of categories will be stored in this file")

    args = parser.parse_args()

    # Assign inputs from command line to variables
    destination = args.destination
    destination_file_path = args.destination_file_path
    source_type = args.source_type
    source_file = args.source_file
    descriptions_file_path = args.descriptions_file_path

    embeddings_model = OpenAIEmbeddings(model="text-embedding-ada-002")
    llm = ChatOpenAI(model_name="gpt-3.5-turbo-0125", temperature=0.7)

    if source_type == "embeddings":
        categories_names, embeddings, embedding_size = load_embeddings_from_file(source_file)
    elif source_type == "descriptions":
        with open(source_file, "r") as f:
            descriptions = json.load(f)
        categories_names = []
        categories_descriptions = []
        for k, v in descriptions.items():
            categories_names.append(k)
            categories_descriptions.append(v)
        embeddings = embeddings_model.embed_documents(categories_descriptions)
        print("Embeddings created")
        embedding_size = len(embeddings[0])
    else:
        with open(source_file, "r") as f:
            categories = json.load(f)
        all_cats = collect_categories(categories["categories"])
        categories_names, categories_descriptions = create_descriptions(all_cats, llm, descriptions_file_path)
        embeddings = embeddings_model.embed_documents(categories_descriptions)
        print("Embeddings created")
        embedding_size = len(embeddings[0])

    if "d" in destination:
        load_to_qdrant(categories_names, embeddings, embedding_size)
        print("Embeddings loaded to Qdrant")
    if "f" in destination:
        print("Embeddings saved to file")
        with open(destination_file_path, "w") as f:
            result = {}
            for cat, e in zip(categories_names, embeddings):
                result[cat] = e
            json.dump(result, f)


if __name__ == "__main__":
    main()
