package ContentHandler

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	ContentRepository "github.com/okanay/backend-holding/repositories/contents"
	"github.com/okanay/backend-holding/services/cache"
	"github.com/okanay/backend-holding/types"
)

const Group = "" // Cache grubu

type Handler struct {
	Repository *ContentRepository.Repository //
	Cache      cache.CacheService            //
}

func NewHandler(pr *ContentRepository.Repository, c cache.CacheService) *Handler { //
	return &Handler{
		Repository: pr,
		Cache:      c,
	}
}

func mapContentToView(pc types.Content) types.ContentView {
	view := types.ContentView{
		ID:          pc.ID,
		Slug:        pc.Slug,
		Identifier:  pc.Identifier,
		Language:    pc.Language,
		Title:       pc.Title,
		Description: pc.Description,
		ImageURL:    pc.ImageURL,
		ContentHTML: pc.ContentHTML,
		Status:      pc.Status,
		CreatedAt:   pc.CreatedAt,
		UpdatedAt:   pc.UpdatedAt,
	}

	if pc.DetailsJSON != nil && *pc.DetailsJSON != "" {
		var detailsMap map[string]any
		if err := json.Unmarshal([]byte(*pc.DetailsJSON), &detailsMap); err == nil {
			view.DetailsJSON = &detailsMap
		}
	}

	if pc.ContentJSON != "" {
		var contentMap map[string]any
		if err := json.Unmarshal([]byte(pc.ContentJSON), &contentMap); err == nil {
			view.ContentJSON = &contentMap
		}
	}
	return view
}

func mapContentsToViews(Contents []types.Content) []types.ContentView {
	views := make([]types.ContentView, 0, len(Contents))
	for _, pc := range Contents {
		views = append(views, mapContentToView(pc))
	}
	return views
}

func filterPublishedContents(contents []types.Content) []types.Content {
	var published []types.Content
	for _, content := range contents {
		if content.Status == types.ContentStatusPublished {
			published = append(published, content)
		}
	}
	return published
}

func findContentByLanguage(contents []types.Content, language string) *types.Content {
	for _, content := range contents {
		if strings.ToLower(content.Language) == language {
			return &content
		}
	}
	return nil
}

func buildAlternateLanguages(contents []types.Content, excludeLang string) []gin.H {
	var alternates []gin.H
	for _, content := range contents {
		if strings.ToLower(content.Language) != excludeLang {
			alternates = append(alternates, gin.H{
				"language": strings.ToLower(content.Language),
				"slug":     content.Slug,
				"title":    content.Title,
			})
		}
	}
	return alternates
}
