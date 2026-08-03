package repository

import (
	"blog-api-dp/internal/model"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCommentRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCommentRepo(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockFn      func()
		comment     *model.Comment
		expectError bool
	}{
		{
			name: "Success - create comment",
			mockFn: func() {
				mock.ExpectQuery(`INSERT INTO comments \(post_id, author_id, content, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id`).
					WithArgs(1, 1, "Test comment", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			comment: &model.Comment{
				PostID:   1,
				AuthorID: 1,
				Content:  "Test comment",
			},
			expectError: false,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectQuery(`INSERT INTO comments \(post_id, author_id, content, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id`).
					WithArgs(1, 1, "Test comment", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			comment: &model.Comment{
				PostID:   1,
				AuthorID: 1,
				Content:  "Test comment",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			err := repo.Create(ctx, tt.comment)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 1, tt.comment.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCommentRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCommentRepo(db)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		mockFn      func()
		id          int
		expected    *model.Comment
		expectError bool
	}{
		{
			name: "Success - get comment by ID",
			mockFn: func() {
				rows := sqlmock.NewRows([]string{"id", "post_id", "author_id", "content", "created_at", "updated_at"}).
					AddRow(1, 1, 1, "Test comment", now, now)
				mock.ExpectQuery(`SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			id: 1,
			expected: &model.Comment{
				ID:        1,
				PostID:    1,
				AuthorID:  1,
				Content:   "Test comment",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectError: false,
		},
		{
			name: "Error - comment not found",
			mockFn: func() {
				mock.ExpectQuery(`SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = \$1`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			id:          999,
			expected:    nil,
			expectError: true,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectQuery(`SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			id:          1,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			comment, err := repo.GetByID(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, comment)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, comment.ID)
				assert.Equal(t, tt.expected.PostID, comment.PostID)
				assert.Equal(t, tt.expected.AuthorID, comment.AuthorID)
				assert.Equal(t, tt.expected.Content, comment.Content)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCommentRepo_GetByPostID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCommentRepo(db)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		mockFn      func()
		postID      int
		limit       int
		offset      int
		expected    []*model.Comment
		expectError bool
	}{
		{
			name: "Success - get comments by post ID",
			mockFn: func() {
				rows := sqlmock.NewRows([]string{"id", "content", "post_id", "author_id", "created_at", "updated_at"}).
					AddRow(1, "Comment 1", 1, 1, now, now).
					AddRow(2, "Comment 2", 1, 2, now.Add(time.Minute), now.Add(time.Minute))
				mock.ExpectQuery(`SELECT id, content, post_id, author_id, created_at, updated_at FROM comments WHERE post_id = \$1 ORDER BY created_at ASC LIMIT \$2 OFFSET \$3`).
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			postID: 1,
			limit:  10,
			offset: 0,
			expected: []*model.Comment{
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
					AuthorID:  2,
					CreatedAt: now.Add(time.Minute),
					UpdatedAt: now.Add(time.Minute),
				},
			},
			expectError: false,
		},
		{
			name: "Success - no comments found",
			mockFn: func() {
				rows := sqlmock.NewRows([]string{"id", "content", "post_id", "author_id", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT id, content, post_id, author_id, created_at, updated_at FROM comments WHERE post_id = \$1 ORDER BY created_at ASC LIMIT \$2 OFFSET \$3`).
					WithArgs(999, 10, 0).
					WillReturnRows(rows)
			},
			postID:      999,
			limit:       10,
			offset:      0,
			expected:    []*model.Comment{},
			expectError: false,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectQuery(`SELECT id, content, post_id, author_id, created_at, updated_at FROM comments WHERE post_id = \$1 ORDER BY created_at ASC LIMIT \$2 OFFSET \$3`).
					WithArgs(1, 10, 0).
					WillReturnError(sql.ErrConnDone)
			},
			postID:      1,
			limit:       10,
			offset:      0,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			comments, err := repo.GetByPostID(ctx, tt.postID, tt.limit, tt.offset)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, comments)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(comments))

				for i, expectedComment := range tt.expected {
					assert.Equal(t, expectedComment.ID, comments[i].ID)
					assert.Equal(t, expectedComment.Content, comments[i].Content)
					assert.Equal(t, expectedComment.PostID, comments[i].PostID)
					assert.Equal(t, expectedComment.AuthorID, comments[i].AuthorID)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCommentRepo_GetCountByPostID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCommentRepo(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockFn      func()
		postID      int
		expected    int
		expectError bool
	}{
		{
			name: "Success - get count by post ID",
			mockFn: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments WHERE post_id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
			},
			postID:      1,
			expected:    5,
			expectError: false,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments WHERE post_id = \$1`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			postID:      1,
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			count, err := repo.GetCountByPostID(ctx, tt.postID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, 0, count)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, count)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCommentRepo_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCommentRepo(db)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		mockFn      func()
		comment     *model.Comment
		expectError bool
	}{
		{
			name: "Success - update comment",
			mockFn: func() {
				mock.ExpectExec(`UPDATE comments SET content = \$1, updated_at = \$2 WHERE id = \$3 AND author_id = \$4`).
					WithArgs("Updated comment", sqlmock.AnyArg(), 1, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			comment: &model.Comment{
				ID:        1,
				Content:   "Updated comment",
				AuthorID:  1,
				UpdatedAt: now,
			},
			expectError: false,
		},
		{
			name: "Error - comment not found",
			mockFn: func() {
				mock.ExpectExec(`UPDATE comments SET content = \$1, updated_at = \$2 WHERE id = \$3 AND author_id = \$4`).
					WithArgs("Updated comment", sqlmock.AnyArg(), 999, 1).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			comment: &model.Comment{
				ID:        999,
				Content:   "Updated comment",
				AuthorID:  1,
				UpdatedAt: now,
			},
			expectError: true,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectExec(`UPDATE comments SET content = \$1, updated_at = \$2 WHERE id = \$3 AND author_id = \$4`).
					WithArgs("Updated comment", sqlmock.AnyArg(), 1, 1).
					WillReturnError(sql.ErrConnDone)
			},
			comment: &model.Comment{
				ID:        1,
				Content:   "Updated comment",
				AuthorID:  1,
				UpdatedAt: now,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			err := repo.Update(ctx, tt.comment)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCommentRepo_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCommentRepo(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockFn      func()
		id          int
		authorID    int
		expectError bool
	}{
		{
			name: "Success - delete comment",
			mockFn: func() {
				mock.ExpectExec(`DELETE FROM comments WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			id: 1,
			//			authorID:    1,
			expectError: false,
		},
		{
			name: "Error - comment not found",
			mockFn: func() {
				mock.ExpectExec(`DELETE FROM comments WHERE id = \$1`).
					WithArgs(99).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			id: 99,
			//			authorID:    1,
			expectError: true,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectExec(`DELETE FROM comments WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			id: 1,
			//	authorID:    1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			err := repo.Delete(ctx, tt.id) // tt.authorID

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
