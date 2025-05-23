package ContentHandler

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	cr "github.com/okanay/backend-holding/repositories/content"
	"github.com/okanay/backend-holding/services/cache"
	"github.com/okanay/backend-holding/types"
)

const Group = cache.GroupContent

type Handler struct {
	Repository *cr.Repository
	Cache      cache.CacheService
}

func NewHandler(repo *cr.Repository, cacheService cache.CacheService) *Handler {
	return &Handler{
		Repository: repo,
		Cache:      cacheService,
	}
}

// mapContentToView - Content'i ContentView'a dönüştürür
func mapContentToView(content types.Content) types.ContentView {
	view := types.ContentView{
		ID:          content.ID,
		Slug:        content.Slug,
		Identifier:  content.Identifier,
		Language:    content.Language,
		Title:       content.Title,
		Description: content.Description,
		Category:    content.Category,
		ImageURL:    content.ImageURL,
		ContentHTML: content.ContentHTML,
		Status:      content.Status,
		CreatedAt:   content.CreatedAt,
		UpdatedAt:   content.UpdatedAt,
	}

	// DetailsJSON dönüşümü
	if content.DetailsJSON != nil && *content.DetailsJSON != "" {
		var detailsMap map[string]any
		if err := json.Unmarshal([]byte(*content.DetailsJSON), &detailsMap); err == nil {
			view.DetailsJSON = &detailsMap
		}
	}

	// ContentJSON dönüşümü
	if content.ContentJSON != "" {
		var contentMap map[string]any
		if err := json.Unmarshal([]byte(content.ContentJSON), &contentMap); err == nil {
			view.ContentJSON = &contentMap
		}
	}

	return view
}

// mapContentsToViews - Content slice'ını ContentView slice'ına dönüştürür
func mapContentsToViews(contents []types.Content) []types.ContentView {
	views := make([]types.ContentView, 0, len(contents))
	for _, content := range contents {
		views = append(views, mapContentToView(content))
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
