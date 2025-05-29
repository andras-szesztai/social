package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/andras-szesztai/social/internal/store"
)

func Seed(store *store.Store, db *sql.DB) error {
	ctx := context.Background()

	users := generateUsers(100)
	tx, _ := db.BeginTx(ctx, nil)

	for _, user := range users {
		_, err := store.Users.Create(ctx, tx, user)
		if err != nil {
			return err
		}
	}

	posts := generatePosts(250, users)
	for _, post := range posts {
		_, err := store.Posts.Create(ctx, post)
		if err != nil {
			return err
		}
	}

	comments := generateComments(500, posts)
	for _, comment := range comments {
		_, err := store.Comments.Create(ctx, comment)
		if err != nil {
			return err
		}
	}

	log.Println("Seed completed")

	return nil
}

var usernames = []string{
	"johndoe",
	"janedoe",
	"jimdoe",
	"jilldoe",
	"jackdoe",
	"robertdoe",
	"marydoe",
	"patriciadoe",
	"lindadoe",
	"barbaradoe",
	"elizabethdoe",
	"jenniferdoe",
}

func generateUsers(count int) []*store.User {
	users := make([]*store.User, count)
	for i := 0; i < count; i++ {
		username := usernames[i%len(usernames)] + fmt.Sprintf(" %d", i+1)
		users[i] = &store.User{
			ID:       int64(i + 1),
			Username: username,
			Email:    fmt.Sprintf("%s@example.com", username),
		}
	}

	return users
}

var postTitles = []string{
	"The best post ever",
	"The worst post ever",
	"The most interesting post ever",
	"The most boring post ever",
	"The most controversial post ever",
	"The most controversial post ever",
}

var postContents = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
	"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
	"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
	"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
}

var tags = []string{
	"tag1",
	"tag2",
	"tag3",
	"tag4",
	"tag5",
}

func generatePosts(count int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, count)
	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]
		posts[i] = &store.Post{
			ID:      int64(i + 1),
			Title:   postTitles[rand.Intn(len(postTitles))] + fmt.Sprintf(" %d", i+1),
			Content: postContents[rand.Intn(len(postContents))] + fmt.Sprintf(" %d", i+1),
			UserID:  user.ID,
			Tags:    []string{tags[rand.Intn(len(tags))] + fmt.Sprintf(" %d", i+1)},
		}
	}
	return posts
}

var commentContents = []string{
	"This is an amazing comment",
	"This is a great comment",
	"This is a good comment",
	"This is a bad comment",
	"This is a terrible comment",
	"This is a comment",
}

func generateComments(count int, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, count)
	for i := 0; i < count; i++ {
		post := posts[rand.Intn(len(posts))]
		comments[i] = &store.Comment{
			ID:      int64(i + 1),
			PostID:  post.ID,
			UserID:  post.UserID,
			Content: commentContents[rand.Intn(len(commentContents))] + fmt.Sprintf(" %d", i+1),
		}
	}
	return comments
}
