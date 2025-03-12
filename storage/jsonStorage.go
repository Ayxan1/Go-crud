package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"example.com/gocrud/models"
)

var (
	ErrNotFound  = errors.New("record not found")
	storageMutex sync.RWMutex
)

const (
	dataDirectory = "./data"
	postsFile     = "posts.json"
)

type PostsStorage struct {
	Posts []models.Post `json:"posts"`
}

func init() {
	// Create data directory if it doesn't exist
	if _, err := os.Stat(dataDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(dataDirectory, 0755)
		if err != nil {
			panic(fmt.Sprintf("failed to create data directory: %v", err))
		}
	}

	// Initialize posts file if it doesn't exist
	postsPath := filepath.Join(dataDirectory, postsFile)
	if _, err := os.Stat(postsPath); os.IsNotExist(err) {
		storage := PostsStorage{Posts: []models.Post{}}
		err = savePostsToFile(storage)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize posts file: %v", err))
		}
	}
}

func getPostsPath() string {
	return filepath.Join(dataDirectory, postsFile)
}

func loadPostsFromFile() (PostsStorage, error) {
	storageMutex.RLock()
	defer storageMutex.RUnlock()

	var storage PostsStorage
	postsPath := getPostsPath()

	data, err := os.ReadFile(postsPath)
	if err != nil {
		return storage, fmt.Errorf("failed to read posts file: %v", err)
	}

	err = json.Unmarshal(data, &storage)
	if err != nil {
		return storage, fmt.Errorf("failed to unmarshal posts data: %v", err)
	}

	return storage, nil
}

func savePostsToFile(storage PostsStorage) error {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	postsPath := getPostsPath()
	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal posts data: %v", err)
	}

	err = os.WriteFile(postsPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write posts data: %v", err)
	}

	return nil
}

func GetAllPosts() ([]models.Post, error) {
	storage, err := loadPostsFromFile()
	if err != nil {
		return nil, err
	}
	return storage.Posts, nil
}

func GetPostByID(id uint) (models.Post, error) {
	storage, err := loadPostsFromFile()
	if err != nil {
		return models.Post{}, err
	}

	for _, post := range storage.Posts {
		if post.ID == id {
			return post, nil
		}
	}

	return models.Post{}, ErrNotFound
}

func CreatePost(post *models.Post) error {
	storage, err := loadPostsFromFile()
	if err != nil {
		return err
	}

	// Set ID and timestamps
	maxID := uint(0)
	for _, p := range storage.Posts {
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	post.ID = maxID + 1
	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	storage.Posts = append(storage.Posts, *post)
	return savePostsToFile(storage)
}

func UpdatePost(post *models.Post) error {
	storage, err := loadPostsFromFile()
	if err != nil {
		return err
	}

	found := false
	for i, p := range storage.Posts {
		if p.ID == post.ID {
			// Keep created_at, update updated_at
			post.CreatedAt = p.CreatedAt
			post.UpdatedAt = time.Now()
			storage.Posts[i] = *post
			found = true
			break
		}
	}

	if !found {
		return ErrNotFound
	}

	return savePostsToFile(storage)
}

func DeletePost(id uint) error {
	storage, err := loadPostsFromFile()
	if err != nil {
		return err
	}

	found := false
	newPosts := []models.Post{}
	for _, p := range storage.Posts {
		if p.ID == id {
			found = true
		} else {
			newPosts = append(newPosts, p)
		}
	}

	if !found {
		return ErrNotFound
	}

	storage.Posts = newPosts
	return savePostsToFile(storage)
}
