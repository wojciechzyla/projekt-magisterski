from abc import ABC, abstractmethod
from terms_processing import Processing, ProcessingQueue, STEMMER
from qdrant_client import QdrantClient
from nltk.tokenize import word_tokenize
from copy import deepcopy
from langchain.prompts import ChatPromptTemplate
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain.chains.summarize import load_summarize_chain
from langchain_core.output_parsers import StrOutputParser
from dotenv import load_dotenv
import random
load_dotenv()


class Extractor(ABC):
    @abstractmethod
    def extract_categories(self, text: str):
        pass


class TermsExtractor(Extractor):
    def __init__(self, terms_to_categories: dict):
        # terms_to_categories has a schema: {"key": ["val1", "val2", ...]}
        self.term_to_cat = terms_to_categories

        # create trie structure
        self.trie = {}
        for i, row in enumerate(terms_to_categories.keys()):
            new_term = row.split(" ")
            if new_term[0] in self.trie:
                self.trie[new_term[0]].append([*new_term[1:]])
            else:
                self.trie[new_term[0]] = [[*new_term[1:]]]

    def __find_terms(self, words):
        terms_found = []
        queue = ProcessingQueue()
        for i, word in enumerate(words):
            if word in self.trie:
                terms_list = deepcopy(self.trie[word])
                for i_res, res in enumerate(terms_list):
                    terms_list[i_res].insert(0, word)
                p = Processing(terms_list)
                queue.add_processing(p)
            terms_found += queue.process_queue(i, word)
        return terms_found

    def extract_categories(self, text: str):
        words = word_tokenize(text)
        words = [STEMMER.stem(w) for w in words]
        terms_found = self.__find_terms(words)
        categories = {}
        for term in terms_found:
            key = " ".join(words[term[0]:term[1] + 1])
            categories[key] = self.term_to_cat[key]
        result = {}
        for v in categories.values():
            for cat in v:
                result[cat] = 0.8
        return result


class SimilarityExtractor(Extractor):
    def __init__(self, embeddings_model, qdrant: QdrantClient, collection_name: str, llm):
        self.qdrant = qdrant
        self.collection_name = collection_name
        self.embeddings_model = embeddings_model
        self.llm = llm

    def __return_query_text(self, text):
        num_tokens = self.llm.get_num_tokens(text)
        if num_tokens > 2000:
            text_splitter = RecursiveCharacterTextSplitter(separators=["\n\n", "\n", "."], chunk_size=500,
                                                           chunk_overlap=350)
            docs = text_splitter.create_documents([text])
            chain = load_summarize_chain(llm=self.llm, chain_type='map_reduce')
            output = chain.run(docs)
        elif num_tokens > 400:
            template = """
            %INSTRUCTIONS:
            Please summarize the following piece of text. Make this summary 5-10 sentences long.

            %TEXT:
            {text}
            """
            prompt = ChatPromptTemplate.from_template(template)
            output_parser = StrOutputParser()
            llm_chain = prompt | self.llm | output_parser
            output = llm_chain.invoke({"text": text})
        else:
            output = text
        return output

    def extract_categories(self, text: str):
        text = self.__return_query_text(text)
        query_res = self.embeddings_model.embed_documents([text])[0]
        hits = self.qdrant.search(
            collection_name=self.collection_name,
            query_vector=query_res,
            limit=5,
        )
        categories_for_api = {}
        for hit in hits:
            categories_for_api[hit.payload["cat"]] = hit.score
        return categories_for_api


class SimilarityAndTermsExtractorWrapper(Extractor):
    def __init__(self, similarity_extractor: SimilarityExtractor, terms_extractor: TermsExtractor):
        self.similarity_extractor = similarity_extractor
        self.terms_extractor = terms_extractor

    def extract_categories(self, text: str):
        similarity_categories = self.similarity_extractor.extract_categories(text)
        terms_categories = self.terms_extractor.extract_categories(text)
        min_satisfaction = min(similarity_categories.values())
        max_satisfaction = max(similarity_categories.values())
        for category in terms_categories.keys():
            similarity_value = random.uniform(min_satisfaction, max_satisfaction)
            terms_categories[category] = similarity_value

        result = {key: max(similarity_categories.get(key, float('-inf')), terms_categories.get(key, float('-inf')))
                 for key in similarity_categories.keys() | terms_categories.keys()}

        return result


class CategoriesExtraction:
    def __init__(self, extractors: [Extractor]):
        self.extractors = extractors

    def extract_categories(self, text):
        categories = {}
        for e in self.extractors:
            cats = e.extract_categories(text)
            for c, v in cats.items():
                if c in categories:
                    if categories[c] < v:
                        categories[c] = v
                else:
                    categories[c] = v
        return categories
