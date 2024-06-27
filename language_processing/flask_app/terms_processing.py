from nltk.stem import PorterStemmer

STEMMER = PorterStemmer()


class Processing:
    def __init__(self, terms):
        self.terms = terms
        self.j = 0
        self.delete = False

    def process(self, word_index, word):
        terms_found = []
        indices_to_delete = []
        for i_term, term in enumerate(self.terms):
            if term[self.j] == word and self.j == len(term) - 1:
                terms_found.append([word_index - self.j, word_index])
                indices_to_delete.append(i_term)
            elif term[self.j] != word or self.j == len(term) - 1:
                indices_to_delete.append(i_term)

        self.terms = list(filter(lambda x: self.terms.index(x) not in indices_to_delete, self.terms))
        if self.terms:
            self.j += 1
        else:
            self.delete = True
        return terms_found


class ProcessingQueue:
    def __init__(self):
        self.queue = []

    def add_processing(self, p: Processing):
        self.queue.append(p)

    def process_queue(self, word_index, word):
        indices_to_delete = []
        terms_found = []
        for i, p in enumerate(self.queue):
            terms_found += p.process(word_index, word)
            if p.delete:
                indices_to_delete.append(i)
        self.queue = list(filter(lambda x: self.queue.index(x) not in indices_to_delete, self.queue))
        return terms_found
