# ask_pubmed
A go program which search the papers from ncbi. Currently only suport search pmc database.

# Getting Start：
https://github.com/FayeZheng0/ask_pubmed.git

go build

./ask_pubmed httpd 5009

# API：
Provide two Get and Post method to query papers，both of them return the same data struct. 
## Get method to search
GET /api/search/:query

This API will return search result. query is required. will return 3 papers by default.


## Post method to search
POST  /api/search

This API will return search result. The payload should be a JSON object containing the following fields:

```json
{
  // the query content, required
  "query": "the request query content",
  // the threshold of search similarity score
  "threshold": 0.3,
  // expected return search amount
  "amount": 5
}
```

### Response

```json
{
  "ts": 1686107732760,
  "data": {
    "resq": {
        // the request query 
      "query": "query",
      // the papers info
      "papers": [
        {
            // the similarity score between paper abstract and request reery
          "search_score": 0.47601396,
          "article_title": "Carcinogenesis: Failure of resolution of inflammation?",
          "abstract": "paper abstract",
          "keywords": [
            "Eicosanoid",
            "Carcinogen",
          ],
          "doi": "10.1016/j.pharmthera.2020.107670",
          "pmc_id": "7470770",
          "pm_id": "32891711",
          "author": "Panigrahy Dipak",
          "url": "https://www.ncbi.nlm.nih.gov/pmc/articles/PMC7470770/",
          "pub_year": "2020"
        },
        {
          "search_score": 0.4270485,
          "article_title": "Mechanisms of obesity- and diabetes mellitus-related pancreatic carcinogenesis: a comprehensive and systematic review",
          "abstract": "paper abstract",
          "keywords": [
            "Cancer microenvironment",
            "Gastrointestinal cancer",
            "Endocrine system and metabolic diseases"
          ],
          "doi": "10.1038/s41392-023-01376-w",
          "pmc_id": "10039087",
          "pm_id": "36964133",
          "author": "Zhao Yupei",
          "url": "https://www.ncbi.nlm.nih.gov/pmc/articles/PMC10039087/",
          "pub_year": "2023"
        }
      ],
      // keywords which AI return as search keywords of NCBI
      "keywords": [
        "(\"bacteria\" and \"cancer\") and (\"pathogenesis\" or \"carcinogenesis\" or \"tumorigenesis\" or \"inflammation\")"
      ],
      "threshold": 0.3,
      "paper_count": 3
    }
  }
}

```