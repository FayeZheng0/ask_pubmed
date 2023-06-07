package botastic

import (
	"context"
	"fmt"
	"time"

	"github.com/FayeZheng0/ask_pubmed/core"
	"github.com/fox-one/pkg/logger"
	"github.com/pandodao/botastic-go"
	"github.com/pandodao/tokenizer-go"
	"github.com/yiplee/go-cache"
)

const (
	StatusInit = iota
	StatusPending
	StatusCompleted
	StatusError
)

type Client struct {
	c             *botastic.Client
	botID         uint64
	conversations *cache.Cache[string, string]
}

func New(endpoint, appID, appSecret string, botID uint64) *Client {
	return &Client{
		c:             botastic.New(appID, appSecret, botastic.WithHost(endpoint)),
		conversations: cache.New[string, string](),
		botID:         botID,
	}
}

func (c *Client) SearchIndexes(ctx context.Context, searchQuery *core.SearchQuery) (*botastic.SearchIndexesResponse, error) {
	s, err := c.c.SearchIndexes(ctx, botastic.SearchIndexesRequest{
		Keywords: searchQuery.UserQuery,
		N:        20,
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Client) CreateIndexes(ctx context.Context, paper core.Paper, query string) error {
	if paper.Abstract == "" {
		return nil
	}
	t := tokenizer.MustCalToken(paper.Abstract)
	if t > 3500 {
		//@TODO: handle the exceed
		return nil
	}

	err := c.c.CreateIndexes(ctx, botastic.CreateIndexesRequest{
		Items: []*botastic.CreateIndexesItem{{
			ObjectID:   paper.PmcId,
			Data:       fmt.Sprintf("[%s] - %s", query, paper.Abstract),
			Properties: "",
			Category:   "plain-text",
		},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetKeywords(ctx context.Context, query, remoteAddr string) (string, error) {
	log := logger.FromContext(ctx).WithField("service", "botasticz.GetKeywords")
	id, ok := c.conversations.Get(remoteAddr)
	if !ok {
		conversation, err := c.c.CreateConversation(ctx, botastic.CreateConversationRequest{
			BotID:        c.botID,
			UserIdentity: remoteAddr,
			Lang:         "en",
		})

		if err != nil {
			log.WithError(err).Println("CreateConversation: failed to create conversation")
			return "", err
		}

		id = conversation.ID
		c.conversations.Set(remoteAddr, id)
	}
	turn, err := c.c.PostToConversation(ctx, botastic.PostToConversationPayloadRequest{
		ConversationID: id,
		Content:        query,
		Category:       "plain-text",
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for err == nil && turn.Status < StatusCompleted {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Second):
			turn, err = c.c.GetConvTurn(ctx, id, turn.ID, false)
		}
	}

	if err != nil {
		log.WithError(err).Println("GetConvTurn: failed to get conversation turn")
		return "", err
	}

	if turn.Status == StatusError {
		return "", fmt.Errorf("ended with status error")
	}

	return turn.Response, nil
}

func (c *Client) Post(ctx context.Context, userID, msg string) (string, error) {
	id, ok := c.conversations.Get(userID)
	if !ok {
		conversation, err := c.c.CreateConversation(ctx, botastic.CreateConversationRequest{
			BotID:        c.botID,
			UserIdentity: userID,
			Lang:         "zh",
		})

		if err != nil {
			return "", err
		}

		id = conversation.ID
		c.conversations.Set(userID, id)
	}

	turn, err := c.c.PostToConversation(ctx, botastic.PostToConversationPayloadRequest{
		ConversationID: id,
		Content:        msg,
		Category:       "plain-text",
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for err == nil && turn.Status < StatusCompleted {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Second):
			turn, err = c.c.GetConvTurn(ctx, id, turn.ID, false)
		}
	}

	if err != nil {
		return "", err
	}

	if turn.Status == StatusError {
		return "", fmt.Errorf("ended with status error")
	}

	return turn.Response, nil
}
