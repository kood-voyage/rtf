package store

import "forum/server/internal/models"

type UserRepository interface {
	Create(*models.User) error
	// FindByID is used to send the client user data
	// (password from db has been sanitized)
	FindByID(string) (*models.User, error)
	// Check can be used to get user data not meant to go the client
	// (password from db has NOT BEEN SANITIZED)
	Check(string) (*models.User, error)
}

type ReactionRepository interface {
	Create(*models.Reaction) error
	AddReactionToParent(post_id string, reaction_id string, user_id string) error
	RemoveFromParent(post_id string, reaction_id string, user_id string) error
	GetUserReactionsToParent(parent_id string, user_id string) (*[]models.Reaction, error)
	GetReactionsToParent(parent_id string) (*[]models.Reaction, error)
	GetAllReactions() (*[]models.Reaction, error)
}

type PostRepository interface {
	Create(*models.Post, []models.Category) error
	GetPost(string) (*models.Post, error)
}

type CategoryRepository interface {
	Create(*models.Category) error
	GetCategory(string) (*models.Category, error)
	GetCategoriesForPosts(postId string) (*[]models.Category, error)
	GetAllCategories() (*[]models.Category, error)
}

type CommentRepository interface {
	Create(*models.Comment) error
	GetComment(id string) (*[]models.Comment, error)
}
