package search

import (
	"context"
	"sort"
	"strings"

	"github.com/FayeZheng0/ask_pubmed/core"
	"github.com/FayeZheng0/ask_pubmed/pubmed"
	"github.com/FayeZheng0/ask_pubmed/service/botastic"
	"github.com/fox-one/pkg/logger"
	bgo "github.com/pandodao/botastic-go"
)

func New(botastics *botastic.Client) *service {
	return &service{
		botastics:    botastics,
		searchQuery:  &core.SearchQuery{},
		searchResult: &core.SearchResult{},
	}
}

type service struct {
	botastics    *botastic.Client
	searchQuery  *core.SearchQuery
	searchResult *core.SearchResult
}

func (s *service) SetSearchParams(ctx context.Context, query, remoteAddr string, amount int, threshold float32) error {
	s.searchQuery.UserQuery = query
	s.searchQuery.RemoteAddr = remoteAddr
	s.searchQuery.ExpectedCount = amount
	s.searchQuery.ScoreThreshold = threshold
	if amount == 0 {
		s.searchQuery.ExpectedCount = 3
	}
	if threshold == 0 {
		s.searchQuery.ScoreThreshold = 0.3
	}

	s.searchResult = &core.SearchResult{}
	s.searchResult.Query = query
	s.searchResult.Threshold = s.searchQuery.ScoreThreshold
	return nil
}

func (s *service) GetSearchResults(ctx context.Context) (*core.SearchResult, error) {
	log := logger.FromContext(ctx).WithField("service", "search.GetSearchResults")
	query, err := s.botastics.GetKeywords(ctx, s.searchQuery.UserQuery, s.searchQuery.RemoteAddr)
	if err != nil {
		log.WithError(err).Println("GetKeywords: failed to get keywords")
		return nil, err
	}
	optKws, err := s.GetSearch(ctx, query)
	if err != nil {
		log.WithError(err).Println("GetSearch: failed to get search")
		return nil, err
	}
	count := 0

	for len(optKws) > 0 {
		if count > 10 {
			continue
		}
		optKws, err = s.GetSearch(ctx, strings.Join(optKws, " "))
		if err != nil {
			log.WithError(err).Println("GetSearch: failed to get search")
			return nil, err
		}
		count++
	}

	s.searchResult = sortSearchResult(s.searchResult, s.searchQuery.ExpectedCount)
	return s.searchResult, nil
}

func sortSearchResult(result *core.SearchResult, count int) *core.SearchResult {
	if count > len(result.Papers) {
		return result
	}
	sort.Slice(result.Papers, func(i, j int) bool {
		return result.Papers[i].SearchScore > result.Papers[j].SearchScore
	})

	topPapers := result.Papers[:count]

	result.Papers = topPapers
	result.PaperCount = count

	return result
}

func (s *service) GetSearch(ctx context.Context, query string) ([]string, error) {
	log := logger.FromContext(ctx).WithField("service", "search.GetSearch")
	s.searchResult.Keywords = append(s.searchResult.Keywords, query)
	papers, err := pubmed.ArticlesToPaper(query, s.searchQuery.ExpectedCount)
	if err != nil {
		log.WithError(err).Println("ArticlesToPaper: failed to get papers")
		return nil, err
	}

	b := pubmed.CheckAbstract(papers)
	if b {
		log.Logger.Println("CheckAbstract: get no papers,try re-query")
		query, err := s.botastics.GetKeywords(ctx, s.searchQuery.UserQuery, s.searchQuery.RemoteAddr)
		if err != nil {
			log.WithError(err).Println("GetSearchResults: failed to get search results")
		}
		return []string{query}, err
	}

	for _, paper := range papers {
		err := s.botastics.CreateIndexes(ctx, paper, s.searchQuery.UserQuery)
		if err != nil {
			log.WithError(err).Println("CreateIndexes: failed to create index")
			return nil, err
		}
	}
	searchIndexRes, err := s.botastics.SearchIndexes(ctx, s.searchQuery)
	if err != nil {
		log.WithError(err).Println("SearchIndexes: failed to search index")
		return nil, err
	}
	return s.UpdateSearchResults(papers, searchIndexRes)
}

func (s *service) UpdateSearchResults(papers []core.Paper, searchIndexRes *bgo.SearchIndexesResponse) ([]string, error) {
	papersMap := map[string]core.Paper{}
	for _, paper := range papers {
		papersMap[paper.PmcId] = paper
	}
	for _, item := range searchIndexRes.Items {
		if paper, ok := papersMap[item.ObjectID]; ok {
			paper.SearchScore = item.Score
			papersMap[item.ObjectID] = paper
			if item.Score > s.searchQuery.ScoreThreshold {
				s.searchResult.Papers = append(s.searchResult.Papers, paper)
				s.searchResult.PaperCount = s.searchResult.PaperCount + 1
			}
		}
	}

	if s.searchQuery.ExpectedCount > s.searchResult.PaperCount {
		var papersOpt []core.Paper
		if s.searchResult.Papers != nil {
			papersOpt = s.searchResult.Papers
		} else {
			for _, paper := range papersMap {
				papersOpt = append(papersOpt, paper)
			}
		}
		return pubmed.OptimizeKeywords(papersOpt), nil
	}
	return nil, nil
}
