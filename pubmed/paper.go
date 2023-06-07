package pubmed

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/FayeZheng0/ask_pubmed/core"
	"github.com/biogo/ncbi"
	"github.com/biogo/ncbi/entrez"
)

type ArticleID struct {
	ID        string `xml:",chardata"`
	PubIDType string `xml:"pub-id-type,attr"`
}

type Article struct {
	ArticleTitle string      `xml:"front>article-meta>title-group>article-title"`
	Abstract     []string    `xml:"front>article-meta>abstract>p"`
	Kwd          []string    `xml:"front>article-meta>kwd-group>kwd"`
	ArticleIDs   []ArticleID `xml:"front>article-meta>article-id"`
	Surname      string      `xml:"front>article-meta>contrib-group>contrib>name>surname"`
	GivenNames   string      `xml:"front>article-meta>contrib-group>contrib>name>given-names"`
	Year         string      `xml:"front>article-meta>pub-date>year"`
}

type Articles struct {
	Articles []Article `xml:"article"`
}

func GetEntrez(db, query, email, rettype, retmode string, retmax int) (io.ReadCloser, error) {
	ncbi.SetTimeout(0)
	h := entrez.History{}
	const (
		tool = "entrez.example"
	)
	var (
		p = &entrez.Parameters{RetMax: retmax, RetType: rettype, RetMode: retmode, Sort: "relevance"}
	)
	_, err := entrez.DoSearch(db, query, p, &h, tool, email)
	if err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(1)
	}
	p.Sort = ""

	return entrez.Fetch(db, p, tool, email, &h)
}

func FetchDetails(query string, queryCount int) (*Articles, error) {
	body, err := GetEntrez("pmc", query, "A.N.Other@example.com", "abstract", "xml", queryCount*2)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	var articles Articles
	err = xml.Unmarshal(data, &articles)
	if err != nil {
		return nil, err
	}
	return &articles, nil
}

func ArticlesToPaper(query string, queryCount int) ([]core.Paper, error) {
	articles, err := FetchDetails(query, queryCount)
	// Handler the entrez return error nothing
	if articles == nil {
		articles, err = FetchDetails(query, queryCount)
	}
	if err != nil {
		return nil, err
	}
	papers := make([]core.Paper, 0, len(articles.Articles))
	for _, article := range articles.Articles {
		paper := &core.Paper{}
		paper.ArticleTitle = article.ArticleTitle
		paper.Abstract = strings.Join(article.Abstract, "\n")
		paper.Abstract = strings.TrimSpace(paper.Abstract)
		paper.Keywords = article.Kwd
		paper.Author = fmt.Sprintf("%s %s", article.Surname, article.GivenNames)

		paper.PubYear = article.Year
		for _, id := range article.ArticleIDs {
			switch id.PubIDType {
			case "doi":
				paper.Doi = id.ID
			case "pmc":
				paper.PmcId = id.ID
			case "pmid":
				paper.PmId = id.ID
			}
		}
		paper.Url = fmt.Sprintf("https://www.ncbi.nlm.nih.gov/pmc/articles/PMC%s/", strings.TrimPrefix(paper.PmcId, "PMC"))
		papers = append(papers, *paper)
	}
	return papers, nil
}

func CheckAbstract(papers []core.Paper) bool {
	var count = 0
	for _, paper := range papers {
		if paper.Abstract == "" {
			count++
		}
	}
	if count == len(papers) {
		return true
	} else {
		return false
	}
}
