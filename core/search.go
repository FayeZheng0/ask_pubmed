package core

import (
	"context"
)

type (
	SearchResult struct {
		Query      string   `db:"query" json:"query"`
		Papers     []Paper  `db:"papers" json:"papers"`
		Keywords   []string `json:"keywords"`
		Threshold  float32  `json:"threshold"`
		PaperCount int      `db:"paper_count" json:"paper_count"`
	}

	Paper struct {
		SearchScore  float32  `json:"search_score"`
		ArticleTitle string   `json:"article_title"`
		Abstract     string   `json:"abstract"`
		Keywords     []string `json:"keywords"`
		Doi          string   `json:"doi"`
		PmcId        string   `json:"pmc_id"`
		PmId         string   `json:"pm_id"`
		Author       string   `json:"author"`
		Url          string   `json:"url"`
		PubYear      string   `json:"pub_year"`
	}

	SearchQuery struct {
		UserQuery      string  `json:"user_query"`
		Query          string  `json:"query"`
		ExpectedCount  int     `json:"expected_count"`
		ScoreThreshold float32 `json:"threshold"`
		RemoteAddr     string  `json:"remote_addr"`
	}

	SearchResultService interface {
		GetSearchResults(ctx context.Context) (*SearchResult, error)
		SetSearchParams(ctx context.Context, query, remoteAddr string, amount int, threshold float32) error
	}
)
