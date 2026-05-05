package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
)

const bookmarkMetaKey = "__bookmarks__"

// Bookmark represents a saved reference to a Vault path.
type Bookmark struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// AddBookmark saves a named bookmark pointing to a Vault path.
func AddBookmark(ctx context.Context, client *api.Client, name, path, note string) error {
	if name == "" {
		return fmt.Errorf("bookmark name must not be empty")
	}
	if path == "" {
		return fmt.Errorf("path must not be empty")
	}

	bookmarks, err := loadBookmarks(ctx, client)
	if err != nil {
		return err
	}

	bookmarks[name] = Bookmark{
		Name:      name,
		Path:      path,
		Note:      note,
		CreatedAt: time.Now().UTC(),
	}

	return saveBookmarks(ctx, client, bookmarks)
}

// RemoveBookmark deletes a named bookmark.
func RemoveBookmark(ctx context.Context, client *api.Client, name string) error {
	if name == "" {
		return fmt.Errorf("bookmark name must not be empty")
	}

	bookmarks, err := loadBookmarks(ctx, client)
	if err != nil {
		return err
	}

	if _, ok := bookmarks[name]; !ok {
		return fmt.Errorf("bookmark %q not found", name)
	}

	delete(bookmarks, name)
	return saveBookmarks(ctx, client, bookmarks)
}

// ListBookmarks returns all saved bookmarks.
func ListBookmarks(ctx context.Context, client *api.Client) ([]Bookmark, error) {
	bookmarks, err := loadBookmarks(ctx, client)
	if err != nil {
		return nil, err
	}

	result := make([]Bookmark, 0, len(bookmarks))
	for _, b := range bookmarks {
		result = append(result, b)
	}
	return result, nil
}

func loadBookmarks(ctx context.Context, client *api.Client) (map[string]Bookmark, error) {
	secret, err := client.Logical().ReadWithContext(ctx, bookmarkMetaKey)
	if err != nil || secret == nil {
		return map[string]Bookmark{}, nil
	}

	raw, ok := secret.Data["data"]
	if !ok {
		return map[string]Bookmark{}, nil
	}

	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var result map[string]Bookmark
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func saveBookmarks(ctx context.Context, client *api.Client, bookmarks map[string]Bookmark) error {
	_, err := client.Logical().WriteWithContext(ctx, bookmarkMetaKey, map[string]interface{}{
		"data": bookmarks,
	})
	return err
}
