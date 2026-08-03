package service

import (
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/repository"
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostService_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postRepo := repository.NewPostRepo(db)
	userRepo := repository.NewUserRepo(db)
	service := NewPostService(postRepo, userRepo, nil)

	ctx := context.Background()

	tests := []struct {
		name          string
		userID        int
		req           *model.PostCreateRequest
		mockSetup     func()
		expectedError bool
	}{
		{
			name:   "Success - create published post",
			userID: 1,
			req: &model.PostCreateRequest{
				Title:   "Published Post",
				Content: "Content",
			},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(regexp.QuoteMeta(repository.PostRepo_createPostQuery)).
					WithArgs("Published Post", "Content", 1, "published", nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name:   "Success - create scheduled post",
			userID: 1,
			req: &model.PostCreateRequest{
				Title:     "Scheduled Post",
				Content:   "Content",
				PublishAt: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(regexp.QuoteMeta(repository.PostRepo_createPostQuery)).
					WithArgs("Scheduled Post", "Content", 1, "draft", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name:   "Success - create post with past publish_at becomes published",
			userID: 1,
			req: &model.PostCreateRequest{
				Title:     "Past Scheduled Post",
				Content:   "Content",
				PublishAt: func() *time.Time { t := time.Now().Add(-1 * time.Hour); return &t }(),
			},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(regexp.QuoteMeta(repository.PostRepo_createPostQuery)).
					WithArgs("Past Scheduled Post", "Content", 1, "published", nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name:   "Error - invalid title (empty)",
			userID: 1,
			req: &model.PostCreateRequest{
				Title:   "",
				Content: "Content",
			},
			mockSetup:     func() {},
			expectedError: true,
		},
		{
			name:   "Error - invalid content (empty)",
			userID: 1,
			req: &model.PostCreateRequest{
				Title:   "Valid Title",
				Content: "",
			},
			mockSetup:     func() {},
			expectedError: true,
		},
		{
			name:   "Error - database error",
			userID: 1,
			req: &model.PostCreateRequest{
				Title:   "Error Post",
				Content: "Content",
			},
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(repository.PostRepo_createPostQuery)).
					WithArgs("Error Post", "Content", 1, "published", nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			post, err := service.Create(ctx, tt.userID, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, post)
				assert.Equal(t, tt.req.Title, post.Title)
				assert.Equal(t, tt.req.Content, post.Content)
				assert.Equal(t, tt.userID, post.AuthorID)
				assert.Greater(t, post.ID, 0)
				assert.WithinDuration(t, time.Now(), post.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), post.UpdatedAt, time.Second)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostService_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postRepo := repository.NewPostRepo(db)
	userRepo := repository.NewUserRepo(db)

	// Для GetByID JWTManager не используется.
	servicePost := NewPostService(postRepo, userRepo, nil)

	ctx := context.Background()
	now := time.Now()

	query := regexp.QuoteMeta(repository.PostRepo_getPostByIDQuery)

	tests := []struct {
		name             string
		id               int
		requestorID      int
		mockSetup        func()
		expectedError    bool
		expectedNotFound bool
	}{
		{
			name:        "Success - get published post",
			id:          1,
			requestorID: 0,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Published Post",
					"Content",
					1,
					"published",
					nil,
					now,
					now,
				)

				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedError:    false,
			expectedNotFound: false,
		},
		{
			name:        "Success - get draft post as author",
			id:          1,
			requestorID: 1,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Draft Post",
					"Content",
					1,
					"draft",
					nil,
					now,
					now,
				)

				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedError:    false,
			expectedNotFound: false,
		},
		{
			name:        "Error - post not found",
			id:          999,
			requestorID: 0,
			mockSetup: func() {
				mock.ExpectQuery(query).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError:    true,
			expectedNotFound: true,
		},
		{
			name:        "Error - database error",
			id:          1,
			requestorID: 0,
			mockSetup: func() {
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError:    true,
			expectedNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			post, err := servicePost.GetByID(ctx, tt.id)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, post)
			} else {
				require.NoError(t, err)
				require.NotNil(t, post)

				require.Equal(t, tt.id, post.ID)
				require.Equal(t, "Content", post.Content)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostService_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postRepo := repository.NewPostRepo(db)
	userRepo := repository.NewUserRepo(db)

	servicePost := NewPostService(postRepo, userRepo, nil)

	ctx := context.Background()
	now := time.Now()

	selectQuery := regexp.QuoteMeta(repository.PostRepo_getllQuery)

	countQuery := regexp.QuoteMeta(
		"SELECT COUNT(*) FROM posts WHERE status = 'published'",
	)

	tests := []struct {
		name          string
		limit         int
		offset        int
		mockSetup     func()
		expectedError bool
		expectedCount int
		expectedPosts []*model.Post
	}{
		{
			name:   "Success - get published posts",
			limit:  10,
			offset: 0,
			mockSetup: func() {
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
						"Post 1",
						"Content 1",
						1,
						"published",
						nil,
						now,
						now,
					).
					AddRow(
						2,
						"Post 2",
						"Content 2",
						1,
						"published",
						nil,
						now.Add(time.Hour),
						now.Add(time.Hour),
					)

				mock.ExpectQuery(selectQuery).
					WithArgs(10, 0).
					WillReturnRows(rows)

				mock.ExpectQuery(countQuery).
					WillReturnRows(
						sqlmock.NewRows([]string{"count"}).AddRow(2),
					)
			},
			expectedError: false,
			expectedCount: 2,
			expectedPosts: []*model.Post{
				{
					ID:       1,
					Title:    "Post 1",
					Content:  "Content 1",
					AuthorID: 1,
				},
				{
					ID:       2,
					Title:    "Post 2",
					Content:  "Content 2",
					AuthorID: 1,
				},
			},
		},
		{
			name:   "Success - limit is capped at 100",
			limit:  150,
			offset: 0,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Post",
					"Content",
					1,
					"published",
					nil,
					now,
					now,
				)

				mock.ExpectQuery(selectQuery).
					WithArgs(100, 0).
					WillReturnRows(rows)

				mock.ExpectQuery(countQuery).
					WillReturnRows(
						sqlmock.NewRows([]string{"count"}).AddRow(1),
					)
			},
			expectedError: false,
			expectedCount: 1,
			expectedPosts: []*model.Post{
				{
					ID:       1,
					Title:    "Post",
					Content:  "Content",
					AuthorID: 1,
				},
			},
		},
		{
			name:   "Success - non-positive limit uses default value",
			limit:  -1,
			offset: 0,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Post",
					"Content",
					1,
					"published",
					nil,
					now,
					now,
				)

				mock.ExpectQuery(selectQuery).
					WithArgs(10, 0).
					WillReturnRows(rows)

				mock.ExpectQuery(countQuery).
					WillReturnRows(
						sqlmock.NewRows([]string{"count"}).AddRow(1),
					)
			},
			expectedError: false,
			expectedCount: 1,
			expectedPosts: []*model.Post{
				{
					ID:       1,
					Title:    "Post",
					Content:  "Content",
					AuthorID: 1,
				},
			},
		},
		{
			name:   "Success - zero limit uses default value",
			limit:  0,
			offset: 0,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Post",
					"Content",
					1,
					"published",
					nil,
					now,
					now,
				)

				mock.ExpectQuery(selectQuery).
					WithArgs(10, 0).
					WillReturnRows(rows)

				mock.ExpectQuery(countQuery).
					WillReturnRows(
						sqlmock.NewRows([]string{"count"}).AddRow(1),
					)
			},
			expectedError: false,
			expectedCount: 1,
			expectedPosts: []*model.Post{
				{
					ID:       1,
					Title:    "Post",
					Content:  "Content",
					AuthorID: 1,
				},
			},
		},
		{
			name:   "Success - negative offset uses zero",
			limit:  10,
			offset: -5,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Post",
					"Content",
					1,
					"published",
					nil,
					now,
					now,
				)

				mock.ExpectQuery(selectQuery).
					WithArgs(10, 0).
					WillReturnRows(rows)

				mock.ExpectQuery(countQuery).
					WillReturnRows(
						sqlmock.NewRows([]string{"count"}).AddRow(1),
					)
			},
			expectedError: false,
			expectedCount: 1,
			expectedPosts: []*model.Post{
				{
					ID:       1,
					Title:    "Post",
					Content:  "Content",
					AuthorID: 1,
				},
			},
		},
		{
			name:   "Success - empty result",
			limit:  10,
			offset: 100,
			mockSetup: func() {
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

				mock.ExpectQuery(selectQuery).
					WithArgs(10, 100).
					WillReturnRows(rows)

				mock.ExpectQuery(countQuery).
					WillReturnRows(
						sqlmock.NewRows([]string{"count"}).AddRow(2),
					)
			},
			expectedError: false,
			expectedCount: 2,
			expectedPosts: []*model.Post{},
		},
		{
			name:   "Error - failed to get posts",
			limit:  10,
			offset: 0,
			mockSetup: func() {
				mock.ExpectQuery(selectQuery).
					WithArgs(10, 0).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
			expectedCount: 0,
			expectedPosts: nil,
		},
		{
			name:   "Error - failed to get total count",
			limit:  10,
			offset: 0,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Post",
					"Content",
					1,
					"published",
					nil,
					now,
					now,
				)

				mock.ExpectQuery(selectQuery).
					WithArgs(10, 0).
					WillReturnRows(rows)

				mock.ExpectQuery(countQuery).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
			expectedCount: 0,
			expectedPosts: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			posts, total, err := servicePost.GetAll(ctx, tt.limit, tt.offset)

			if tt.expectedError {
				require.Error(t, err)
				require.Zero(t, total)
				require.Nil(t, posts)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedCount, total)
				require.Len(t, posts, len(tt.expectedPosts))

				for i, expectedPost := range tt.expectedPosts {
					require.Equal(t, expectedPost.ID, posts[i].ID)
					require.Equal(t, expectedPost.Title, posts[i].Title)
					require.Equal(t, expectedPost.Content, posts[i].Content)
					require.Equal(t, expectedPost.AuthorID, posts[i].AuthorID)
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostService_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		id            int
		userID        int
		req           *model.PostUpdateRequest
		mockSetup     func(mock sqlmock.Sqlmock, now time.Time)
		expectedError bool
	}{
		{
			name:   "success - update post",
			id:     1,
			userID: 1,
			req: &model.PostUpdateRequest{
				Title:   "Updated Title",
				Content: "Updated Content",
			},
			mockSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				selectPostQuery := regexp.QuoteMeta(`
     SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
     FROM posts
     WHERE id = $1
    `)

				updatePostQuery := regexp.QuoteMeta(`
     UPDATE posts
     SET title = $1, content = $2, updated_at = $3
     WHERE id = $4
    `)

				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Old Title",
					"Old Content",
					1,
					"published",
					nil,
					now,
					now,
				)

				mock.ExpectQuery(selectPostQuery).
					WithArgs(1).
					WillReturnRows(rows)

				mock.ExpectExec(updatePostQuery).
					WithArgs(
						"Updated Title",
						"Updated Content",
						sqlmock.AnyArg(),
						1,
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			t.Cleanup(func() {
				_ = db.Close()
			})

			postRepo := repository.NewPostRepo(db)
			userRepo := repository.NewUserRepo(db)
			servicePost := NewPostService(postRepo, userRepo, nil)

			now := time.Now()
			tt.mockSetup(mock, now)

			updatedPost, err := servicePost.Update(
				ctx,
				tt.id,
				tt.userID,
				tt.req,
			)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, updatedPost)
			} else {
				require.NoError(t, err)
				require.NotNil(t, updatedPost)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostService_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postRepo := repository.NewPostRepo(db)
	userRepo := repository.NewUserRepo(db)
	servicePost := NewPostService(postRepo, userRepo, nil)

	ctx := context.Background()

	selectQuery := regexp.QuoteMeta(
		`SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
   FROM posts
   WHERE id = $1`,
	)

	deleteQuery := regexp.QuoteMeta(
		`DELETE FROM posts WHERE id = $1`,
	)

	tests := []struct {
		name          string
		id            int
		userID        int
		postAuthorID  int
		mockSetup     func()
		expectedError bool
	}{
		{
			name:         "Success - delete post",
			id:           1,
			userID:       1,
			postAuthorID: 1,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Post",
					"Content",
					1,
					"published",
					nil,
					time.Now(),
					time.Now(),
				)

				mock.ExpectQuery(selectQuery).
					WithArgs(1).
					WillReturnRows(rows)

				mock.ExpectExec(deleteQuery).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name:         "Error - post not found",
			id:           999,
			userID:       1,
			postAuthorID: 1,
			mockSetup: func() {
				mock.ExpectQuery(selectQuery).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: true,
		},
		{
			name:         "Error - unauthorized user",
			id:           1,
			userID:       2,
			postAuthorID: 1,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Post",
					"Content",
					1,
					"published",
					nil,
					time.Now(),
					time.Now(),
				)

				mock.ExpectQuery(selectQuery).
					WithArgs(1).
					WillReturnRows(rows)

				// DELETE не должен выполняться,
				// поскольку пользователь не является автором.
			},
			expectedError: true,
		},
		{
			name:         "Error - database delete error",
			id:           1,
			userID:       1,
			postAuthorID: 1,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"id",
					"title",
					"content",
					"author_id",
					"status",
					"publish_at",
					"created_at",
					"updated_at",
				}).AddRow(
					1,
					"Post",
					"Content",
					1,
					"published",
					nil,
					time.Now(),
					time.Now(),
				)

				mock.ExpectQuery(selectQuery).
					WithArgs(1).
					WillReturnRows(rows)

				mock.ExpectExec(deleteQuery).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := servicePost.Delete(ctx, tt.id, tt.userID)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostService_GetScheduledPosts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postRepo := repository.NewPostRepo(db)
	userRepo := repository.NewUserRepo(db)
	servicePost := NewPostService(postRepo, userRepo, nil)

	ctx := context.Background()
	now := time.Now()

	scheduledAt1 := now
	scheduledAt2 := now.Add(time.Hour)

	query := regexp.QuoteMeta(`
  SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
  FROM posts
  WHERE status = 'draft'
   AND publish_at IS NOT NULL
   AND publish_at <= NOW()
  ORDER BY publish_at ASC
 `)

	tests := []struct {
		name           string
		mockSetup      func()
		expectedError  bool
		expectedOutput []*model.Post
	}{
		{
			name: "Success - get scheduled posts",
			mockSetup: func() {
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
						scheduledAt1,
						now,
						now,
					).
					AddRow(
						2,
						"Another Post",
						"Content 2",
						1,
						"draft",
						scheduledAt2,
						now,
						now,
					)

				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
			expectedOutput: []*model.Post{
				{
					ID:        1,
					Title:     "Scheduled Post",
					Content:   "Content",
					AuthorID:  1,
					Status:    "draft",
					PublishAt: &scheduledAt1,
				},
				{
					ID:        2,
					Title:     "Another Post",
					Content:   "Content 2",
					AuthorID:  1,
					Status:    "draft",
					PublishAt: &scheduledAt2,
				},
			},
		},
		{
			name: "Success - no scheduled posts",
			mockSetup: func() {
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
			expectedOutput: []*model.Post{},
		},
		{
			name: "Error - database error",
			mockSetup: func() {
				mock.ExpectQuery(query).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			posts, err := servicePost.GetScheduledPosts(ctx)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, posts)
			} else {
				require.NoError(t, err)
				require.Len(t, posts, len(tt.expectedOutput))

				for i, expected := range tt.expectedOutput {
					actual := posts[i]

					require.Equal(t, expected.ID, actual.ID)
					require.Equal(t, expected.Title, actual.Title)
					require.Equal(t, expected.Content, actual.Content)
					require.Equal(t, expected.AuthorID, actual.AuthorID)
					require.Equal(t, expected.Status, actual.Status)

					if expected.PublishAt == nil {
						require.Nil(t, actual.PublishAt)
					} else {
						require.NotNil(t, actual.PublishAt)
						require.WithinDuration(
							t,
							*expected.PublishAt,
							*actual.PublishAt,
							time.Second,
						)
					}
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostService_GetByAuthor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postRepo := repository.NewPostRepo(db)
	userRepo := repository.NewUserRepo(db)
	servicePost := NewPostService(postRepo, userRepo, nil)

	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name string

		authorID int
		limit    int
		offset   int

		mockSetup func()

		expectedPostsLen int
		expectedCount    int
		expectedError    bool
	}{
		{
			name:     "success - get posts by author",
			authorID: 1,
			limit:    10,
			offset:   0,
			mockSetup: func() {
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
						"Post 1",
						"Content 1",
						1,
						"published",
						now,
						now,
						now,
					).
					AddRow(
						2,
						"Post 2",
						"Content 2",
						1,
						"published",
						now,
						now,
						now,
					)

				mock.ExpectQuery(".*").
					WithArgs(1, 10, 0).
					WillReturnRows(rows)

				mock.ExpectQuery(".*").
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{"count"}).
							AddRow(2),
					)
			},
			expectedPostsLen: 2,
			expectedCount:    2,
			expectedError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			posts, count, err := servicePost.GetByAuthor(
				ctx,
				tt.authorID,
				tt.limit,
				tt.offset,
			)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, posts)
				require.Zero(t, count)
				require.NoError(t, mock.ExpectationsWereMet())

				return
			}

			require.NoError(t, err)
			require.Len(t, posts, tt.expectedPostsLen)
			require.Equal(t, tt.expectedCount, count)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func timePtr(value time.Time) *time.Time {
	return &value
}
