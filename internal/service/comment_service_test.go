package service

import (
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/repository"
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentService_GetByID(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		id             int
		mockSetup      func(sqlmock.Sqlmock)
		expectedError  bool
		expectedOutput *model.Comment
	}{
		{
			name: "Success - comment found",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id",
					"post_id",
					"author_id",
					"content",
					"created_at",
					"updated_at",
				}).
					AddRow(
						1,
						1,
						1,
						"Found comment",
						now,
						now,
					)

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = $1",
				)).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedOutput: &model.Comment{
				ID:        1,
				PostID:    1,
				AuthorID:  1,
				Content:   "Found comment",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "Error - comment not found",
			id:   999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = $1",
				)).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError:  true,
			expectedOutput: nil,
		},
		{
			name: "Error - database error",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = $1",
				)).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError:  true,
			expectedOutput: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repoCmnt := repository.NewCommentRepo(db)
			repoPost := repository.NewPostRepo(db)
			repoUser := repository.NewUserRepo(db)

			serviceComm := NewCommentService(repoCmnt, repoPost, repoUser)

			tt.mockSetup(mock)

			comment, err := serviceComm.GetByID(
				context.Background(),
				tt.id,
			)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, comment)
			} else {
				require.NoError(t, err)
				require.NotNil(t, comment)

				assert.Equal(t, tt.expectedOutput.ID, comment.ID)
				assert.Equal(t, tt.expectedOutput.PostID, comment.PostID)
				assert.Equal(t, tt.expectedOutput.AuthorID, comment.AuthorID)
				assert.Equal(t, tt.expectedOutput.Content, comment.Content)
				assert.Equal(t, tt.expectedOutput.CreatedAt, comment.CreatedAt)
				assert.Equal(t, tt.expectedOutput.UpdatedAt, comment.UpdatedAt)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Получить комментарии к посту с пагинацией
func TestCommentService_GetByPost(t *testing.T) {
	now := time.Now()

	expectedComments := []*model.Comment{
		{
			ID:        1,
			Content:   "Comment 1",
			PostID:    1,
			AuthorID:  1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        2,
			Content:   "Comment 2",
			PostID:    1,
			AuthorID:  1,
			CreatedAt: now.Add(time.Minute),
			UpdatedAt: now.Add(time.Minute),
		},
	}

	tests := []struct {
		name           string
		postID         int
		offset         int
		limit          int
		expectedOutput []*model.Comment
		expectedCount  int
		expectedError  bool
		mockSetup      func(mock sqlmock.Sqlmock)
	}{
		{
			name:           "success",
			postID:         1,
			offset:         1,
			limit:          10,
			expectedOutput: expectedComments,
			expectedCount:  2,
			expectedError:  false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(".*").
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{"exists"}).
							AddRow(true),
					)

				rows := sqlmock.NewRows([]string{
					"id",
					"content",
					"post_id",
					"author_id",
					"created_at",
					"updated_at",
				}).
					AddRow(
						1,
						"Comment 1",
						1,
						1,
						now,
						now,
					).
					AddRow(
						2,
						"Comment 2",
						1,
						1,
						now.Add(time.Minute),
						now.Add(time.Minute),
					)

				mock.ExpectQuery(".*").
					WithArgs(1, 10, 1).
					WillReturnRows(rows)

				mock.ExpectQuery(".*").
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{"count"}).
							AddRow(2),
					)
			},
		},
		{
			name:           "post not found",
			postID:         999,
			offset:         1,
			limit:          10,
			expectedOutput: nil,
			expectedCount:  0,
			expectedError:  true,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(".*").
					WithArgs(999).
					WillReturnRows(
						sqlmock.NewRows([]string{"exists"}).
							AddRow(false),
					)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			defer func() {
				_ = db.Close()
			}()

			tt.mockSetup(mock)

			repoCmnt := repository.NewCommentRepo(db)
			repoPost := repository.NewPostRepo(db)
			repoUser := repository.NewUserRepo(db)

			serviceComm := NewCommentService(repoCmnt, repoPost, repoUser)

			gotComments, gotCount, err := serviceComm.GetByPost(
				context.Background(),
				tt.postID,
				tt.limit,
				tt.offset,
			)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, gotComments)
				assert.Equal(t, tt.expectedCount, gotCount)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, gotComments)
				assert.Equal(t, tt.expectedCount, gotCount)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCommentService_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repoCmnt := repository.NewCommentRepo(db)
	repoPost := repository.NewPostRepo(db)
	repoUser := repository.NewUserRepo(db)

	serviceComm := NewCommentService(repoCmnt, repoPost, repoUser)

	ctx := context.Background()

	const (
		validPostID  = 1
		validUserID  = 1
		validContent = "This is a comment!"
	)

	tests := []struct {
		name           string
		userID         int
		req            *model.CommentCreateRequest
		mockSetup      func()
		expectedError  bool
		expectedOutput *model.Comment
	}{
		{
			name:   "Success - create comment",
			userID: validUserID,
			req: &model.CommentCreateRequest{
				PostID:  validPostID,
				Content: validContent,
			},
			mockSetup: func() {
				// Сервис сначала проверяет существование поста.
				expectPostExists(mock, validPostID, true)

				// Затем создаёт комментарий.
				mock.ExpectQuery(".*").
					WithArgs(
						validPostID,
						validUserID,
						validContent,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnRows(
						sqlmock.NewRows([]string{"id"}).AddRow(1),
					)
			},
			expectedError: false,
			expectedOutput: &model.Comment{
				ID:       1,
				PostID:   validPostID,
				AuthorID: validUserID,
				Content:  validContent,
			},
		},
		{
			name:   "Error - invalid post ID",
			userID: validUserID,
			req: &model.CommentCreateRequest{
				PostID:  0,
				Content: validContent,
			},
			mockSetup: func() {
				// По логам видно, что сервис проверяет пост даже с ID 0.
				expectPostExists(mock, 0, false)
			},
			expectedError: true,
		},
		{
			name:   "Error - empty content",
			userID: validUserID,
			req: &model.CommentCreateRequest{
				PostID:  validPostID,
				Content: "",
			},
			mockSetup: func() {
				// Валидация контента завершается до обращения к БД.
			},
			expectedError: true,
		},
		{
			name:   "Error - content too long",
			userID: validUserID,
			req: &model.CommentCreateRequest{
				PostID:  validPostID,
				Content: string(make([]byte, 1001)),
			},
			mockSetup: func() {
				// Валидация контента завершается до обращения к БД.
			},
			expectedError: true,
		},
		{
			name:   "Error - database error",
			userID: validUserID,
			req: &model.CommentCreateRequest{
				PostID:  validPostID,
				Content: validContent,
			},
			mockSetup: func() {
				// Сначала проверка существования поста.
				expectPostExists(mock, validPostID, true)

				// Затем ошибка при создании комментария.
				mock.ExpectQuery(".*").
					WithArgs(
						validPostID,
						validUserID,
						validContent,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			comment, err := serviceComm.Create(
				ctx,
				tt.userID,
				tt.req,
			)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, comment)
			} else {
				require.NoError(t, err)
				require.NotNil(t, comment)

				assert.Equal(
					t,
					tt.expectedOutput.ID,
					comment.ID,
				)
				assert.Equal(
					t,
					tt.expectedOutput.PostID,
					comment.PostID,
				)
				assert.Equal(
					t,
					tt.expectedOutput.AuthorID,
					comment.AuthorID,
				)
				assert.Equal(
					t,
					tt.expectedOutput.Content,
					comment.Content,
				)

				assert.False(t, comment.CreatedAt.IsZero())
				assert.False(t, comment.UpdatedAt.IsZero())
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

////

func TestCommentService_UpdateComment(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		req           *model.CommentUpdateRequest
		authorID      int
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name: "Success - update comment",
			id:   1,
			req: &model.CommentUpdateRequest{
				Content: "Updated content",
			},
			authorID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					regexp.QuoteMeta(
						"SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = $1",
					),
				).
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id",
							"post_id",
							"author_id",
							"content",
							"created_at",
							"updated_at",
						}).AddRow(
							1,
							1,
							1,
							"Old content",
							time.Now(),
							time.Now(),
						),
					)

				mock.ExpectExec(
					regexp.QuoteMeta(
						"UPDATE comments SET content = $1, updated_at = $2 WHERE id = $3 AND author_id = $4",
					),
				).
					WithArgs(
						"Updated content",
						sqlmock.AnyArg(),
						1,
						1,
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name: "Error - empty content",
			id:   1,
			req: &model.CommentUpdateRequest{
				Content: "",
			},
			authorID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				// При невалидном content запросы к БД не выполняются.
			},
			expectedError: true,
		},
		{
			name: "Error - content too long",
			id:   1,
			req: &model.CommentUpdateRequest{
				Content: string(make([]byte, 1001)),
			},
			authorID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				// При слишком длинном content запросы к БД не выполняются.
			},
			expectedError: true,
		},
		{
			name: "Error - comment not found",
			id:   999,
			req: &model.CommentUpdateRequest{
				Content: "Updated content",
			},
			authorID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					regexp.QuoteMeta(
						"SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = $1",
					),
				).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)

				// UPDATE не ожидается, потому что комментарий не найден.
			},
			expectedError: true,
		},
		{
			name: "Error - unauthorized author",
			id:   1,
			req: &model.CommentUpdateRequest{
				Content: "Updated content",
			},
			authorID: 2,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					regexp.QuoteMeta(
						"SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = $1",
					),
				).
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id",
							"post_id",
							"author_id",
							"content",
							"created_at",
							"updated_at",
						}).AddRow(
							1,
							1,
							1,
							"Old content",
							time.Now(),
							time.Now(),
						),
					)

				// UPDATE не ожидается, потому что authorID не совпадает.
			},
			expectedError: true,
		},
		{
			name: "Error - database update error",
			id:   1,
			req: &model.CommentUpdateRequest{
				Content: "Updated content",
			},
			authorID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					regexp.QuoteMeta(
						"SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = $1",
					),
				).
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id",
							"post_id",
							"author_id",
							"content",
							"created_at",
							"updated_at",
						}).AddRow(
							1,
							1,
							1,
							"Old content",
							time.Now(),
							time.Now(),
						),
					)

				mock.ExpectExec(
					regexp.QuoteMeta(
						"UPDATE comments SET content = $1, updated_at = $2 WHERE id = $3 AND author_id = $4",
					),
				).
					WithArgs(
						"Updated content",
						sqlmock.AnyArg(),
						1,
						1,
					).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repoCmnt := repository.NewCommentRepo(db)
			repoPost := repository.NewPostRepo(db)
			repoUser := repository.NewUserRepo(db)

			service := NewCommentService(repoCmnt, repoPost, repoUser)

			tt.mockSetup(mock)

			comment, err := service.Update(
				context.Background(),
				tt.id,
				tt.authorID,
				tt.req,
			)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, comment)
			} else {
				require.NoError(t, err)
				require.NotNil(t, comment)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCommentService_DeleteComment(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		authorID      int
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name:     "Success - delete comment",
			id:       1,
			authorID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					regexp.QuoteMeta(`
                        SELECT id, post_id, author_id, content, created_at, updated_at
                        FROM comments
                        WHERE id = $1
                    `),
				).
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id",
							"post_id",
							"author_id",
							"content",
							"created_at",
							"updated_at",
						}).AddRow(
							1,
							1,
							1,
							"content to del",
							time.Now(),
							time.Now(),
						),
					)

				mock.ExpectExec(
					regexp.QuoteMeta(`DELETE FROM comments WHERE id = $1`),
				).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:          "Error - comment not found",
			id:            999,
			authorID:      1,
			expectedError: true,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					regexp.QuoteMeta(`
                        SELECT id, post_id, author_id, content, created_at, updated_at
                        FROM comments
                        WHERE id = $1
                    `),
				).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)

				// DELETE здесь не ожидается.
			},
		},
		{
			name:          "Error - unauthorized delete",
			id:            1,
			authorID:      999,
			expectedError: true,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					regexp.QuoteMeta(`
                        SELECT id, post_id, author_id, content, created_at, updated_at
                        FROM comments
                        WHERE id = $1
                    `),
				).
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id",
							"post_id",
							"author_id",
							"content",
							"created_at",
							"updated_at",
						}).AddRow(
							1,
							1,
							1,
							"content to del",
							time.Now(),
							time.Now(),
						),
					)

				// Автор не совпадает, поэтому DELETE не ожидается.
			},
		},
		{
			name:          "Error - database error",
			id:            1,
			authorID:      1,
			expectedError: true,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					regexp.QuoteMeta(`
                        SELECT id, post_id, author_id, content, created_at, updated_at
                        FROM comments
                        WHERE id = $1
                    `),
				).
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id",
							"post_id",
							"author_id",
							"content",
							"created_at",
							"updated_at",
						}).AddRow(
							1,
							1,
							1,
							"content to del",
							time.Now(),
							time.Now(),
						),
					)

				mock.ExpectExec(
					regexp.QuoteMeta(`DELETE FROM comments WHERE id = $1`),
				).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repoCmnt := repository.NewCommentRepo(db)
			repoPost := repository.NewPostRepo(db)
			repoUser := repository.NewUserRepo(db)

			service := NewCommentService(repoCmnt, repoPost, repoUser)

			tt.mockSetup(mock)

			err = service.Delete(
				context.Background(),
				tt.id,
				tt.authorID,
			)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func expectPostExists(mock sqlmock.Sqlmock, postID int, exists bool) {
	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)",
		),
	).
		WithArgs(postID).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).AddRow(exists),
		)
}
