package repository

import (
	"blog-api-dp/internal/model"
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// test for GetScheduledPosts
func TestPostRepo_GetScheduledPosts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostRepo(db)
	ctx := context.Background()
	now := time.Now()

	query := regexp.QuoteMeta(`
  SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
  FROM posts
  WHERE status = 'draft'
   AND publish_at IS NOT NULL
   AND publish_at <= NOW()
  ORDER BY publish_at ASC
 `)

	tests := []struct {
		name          string
		mockFn        func()
		expectedPosts []*model.Post
		expectError   bool
	}{
		{
			name: "Success - get scheduled posts",
			mockFn: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).
					AddRow(
						1,
						"Scheduled Post",
						"Content",
						1,
						"draft",
						now,
						now,
						now,
					).
					AddRow(
						2,
						"Another Post",
						"Content 2",
						1,
						"draft",
						now.Add(time.Hour),
						now,
						now,
					)

				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
			expectedPosts: []*model.Post{
				{
					ID:        1,
					Title:     "Scheduled Post",
					Content:   "Content",
					AuthorID:  1,
					Status:    "draft",
					PublishAt: &now,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:       2,
					Title:    "Another Post",
					Content:  "Content 2",
					AuthorID: 1,
					Status:   "draft",
					PublishAt: func() *time.Time {
						publishAt := now.Add(time.Hour)
						return &publishAt
					}(),
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectError: false,
		},
		{
			name: "Success - no scheduled posts",
			mockFn: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				})

				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
			// В реализации используется:
			// var posts []*model.Post
			// Поэтому при отсутствии строк возвращается nil.
			expectedPosts: nil,
			expectError:   false,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectQuery(query).
					WillReturnError(sql.ErrConnDone)
			},
			expectedPosts: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			posts, err := repo.GetScheduledPosts(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, posts)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPosts, posts)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// test for PublishPost
func TestPostRepo_PublishPost(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostRepo(db)

	ctx := context.Background()
	postID := 1

	tests := []struct {
		name        string
		mockFn      func()
		expectError bool
	}{
		{
			name: "Success - publish post",
			mockFn: func() {
				mock.ExpectExec(`UPDATE posts SET status = 'published', publish_at = NULL, updated_at = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), postID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Error - post not found",
			mockFn: func() {
				mock.ExpectExec(`UPDATE posts SET status = 'published', publish_at = NULL, updated_at = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), postID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectError: true,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectExec(`UPDATE posts SET status = 'published', publish_at = NULL, updated_at = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), postID).
					WillReturnError(sql.ErrConnDone)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			err := repo.PublishPost(ctx, postID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// test for Create with publish_at
func TestPostRepo_CreateWithPublishAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostRepo(db)

	ctx := context.Background()
	now := time.Now()
	publishAt := now.Add(24 * time.Hour)

	post := &model.Post{
		Title:     "Scheduled Post",
		Content:   "Content",
		AuthorID:  1,
		Status:    "draft",
		PublishAt: &publishAt,
	}

	t.Run("Success - create post with publish_at", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(PostRepo_createPostQuery)).
			WithArgs(
				post.Title,
				post.Content,
				post.AuthorID,
				post.Status,
				publishAt,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).AddRow(1),
			)

		err := repo.Create(ctx, post)

		assert.NoError(t, err)
		//assert.Equal(t, int64(1), post.ID)
		assert.Equal(t, 1, post.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
