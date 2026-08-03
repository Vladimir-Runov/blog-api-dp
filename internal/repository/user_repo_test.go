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

// TestUserRepo_Create tests the Create method of UserRepo.
func TestUserRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockFn      func()
		user        *model.User
		expectError bool
	}{
		{
			name: "Success - create user",
			mockFn: func() {
				mock.ExpectQuery(`INSERT INTO users \(username, email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id`).
					WithArgs("testuser", "test@example.com", "hashedpassword", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "hashedpassword",
			},
			expectError: false,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectQuery(`INSERT INTO users \(username, email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id`).
					WithArgs("testuser", "test@example.com", "hashedpassword", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "hashedpassword",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			err := repo.Create(ctx, tt.user)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 1, tt.user.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		mockFn      func()
		id          int
		expected    *model.User
		expectError bool
	}{
		{
			name: "Success - get user by ID",
			mockFn: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "created_at", "updated_at"}).
					AddRow(1, "testuser", "test@example.com", "hashedpassword", now, now)
				mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			id: 1,
			expected: &model.User{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				Password:  "hashedpassword",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectError: false,
		},
		{
			name: "Error - user not found",
			mockFn: func() {
				mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = \$1`).
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
				mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = \$1`).
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

			user, err := repo.GetByID(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, user.ID)
				assert.Equal(t, tt.expected.Username, user.Username)
				assert.Equal(t, tt.expected.Email, user.Email)
				assert.Equal(t, tt.expected.Password, user.Password)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		mockFn      func()
		email       string
		expected    *model.User
		expectError bool
	}{
		{
			name: "Success - get user by email",
			mockFn: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "created_at", "updated_at"}).
					AddRow(1, "testuser", "test@example.com", "hashedpassword", now, now)
				mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			email: "test@example.com",
			expected: &model.User{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				Password:  "hashedpassword",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectError: false,
		},
		{
			name: "Error - user not found",
			mockFn: func() {
				mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			email:       "nonexistent@example.com",
			expected:    nil,
			expectError: true,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnError(sql.ErrConnDone)
			},
			email:       "test@example.com",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			user, err := repo.GetByEmail(ctx, tt.email)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, user.ID)
				assert.Equal(t, tt.expected.Username, user.Username)
				assert.Equal(t, tt.expected.Email, user.Email)
				assert.Equal(t, tt.expected.Password, user.Password)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		mockFn      func()
		username    string
		expected    *model.User
		expectError bool
	}{
		{
			name: "Success - get user by username",
			mockFn: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "created_at", "updated_at"}).
					AddRow(1, "testuser", "test@example.com", "hashedpassword", now, now)
				mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE username = \$1`).
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			username: "testuser",
			expected: &model.User{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				Password:  "hashedpassword",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectError: false,
		},
		{
			name: "Error - user not found",
			mockFn: func() {
				mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE username = \$1`).
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			username:    "nonexistent",
			expected:    nil,
			expectError: true,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE username = \$1`).
					WithArgs("testuser").
					WillReturnError(sql.ErrConnDone)
			},
			username:    "testuser",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			user, err := repo.GetByUsername(ctx, tt.username)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, user.ID)
				assert.Equal(t, tt.expected.Username, user.Username)
				assert.Equal(t, tt.expected.Email, user.Email)
				assert.Equal(t, tt.expected.Password, user.Password)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_ExistsByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockFn      func()
		email       string
		expected    bool
		expectError bool
	}{
		{
			name: "Success - email exists",
			mockFn: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			email:       "test@example.com",
			expected:    true,
			expectError: false,
		},
		{
			name: "Success - email does not exist",
			mockFn: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1\)`).
					WithArgs("nonexistent@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			email:       "nonexistent@example.com",
			expected:    false,
			expectError: false,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnError(sql.ErrConnDone)
			},
			email:       "test@example.com",
			expected:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			exists, err := repo.ExistsByEmail(ctx, tt.email)

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, exists)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, exists)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_ExistsByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockFn      func()
		username    string
		expected    bool
		expectError bool
	}{
		{
			name: "Success - username exists",
			mockFn: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE username = \$1\)`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			username:    "testuser",
			expected:    true,
			expectError: false,
		},
		{
			name: "Success - username does not exist",
			mockFn: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE username = \$1\)`).
					WithArgs("nonexistent").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			username:    "nonexistent",
			expected:    false,
			expectError: false,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE username = \$1\)`).
					WithArgs("testuser").
					WillReturnError(sql.ErrConnDone)
			},
			username:    "testuser",
			expected:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			exists, err := repo.ExistsByUsername(ctx, tt.username)

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, exists)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, exists)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		mockFn      func()
		user        *model.User
		expectError bool
	}{
		{
			name: "Success - update user",
			mockFn: func() {
				mock.ExpectExec(`UPDATE users SET username = \$1, email = \$2, updated_at = \$3 WHERE id = \$4`).
					WithArgs("updateduser", "updated@example.com", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			user: &model.User{
				ID:        1,
				Username:  "updateduser",
				Email:     "updated@example.com",
				UpdatedAt: now,
			},
			expectError: false,
		},
		{
			name: "Error - user not found",
			mockFn: func() {
				mock.ExpectExec(`UPDATE users SET username = \$1, email = \$2, updated_at = \$3 WHERE id = \$4`).
					WithArgs("updateduser", "updated@example.com", sqlmock.AnyArg(), 999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			user: &model.User{
				ID:        999,
				Username:  "updateduser",
				Email:     "updated@example.com",
				UpdatedAt: now,
			},
			expectError: true,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectExec(`UPDATE users SET username = \$1, email = \$2, updated_at = \$3 WHERE id = \$4`).
					WithArgs("updateduser", "updated@example.com", sqlmock.AnyArg(), 1).
					WillReturnError(sql.ErrConnDone)
			},
			user: &model.User{
				ID:        1,
				Username:  "updateduser",
				Email:     "updated@example.com",
				UpdatedAt: now,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			err := repo.Update(ctx, tt.user)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockFn      func()
		id          int
		expectError bool
	}{
		{
			name: "Success - delete user",
			mockFn: func() {
				mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			id:          1,
			expectError: false,
		},
		{
			name: "Error - user not found",
			mockFn: func() {
				mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			id:          999,
			expectError: true,
		},
		{
			name: "Error - database error",
			mockFn: func() {
				mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			id:          1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()

			err := repo.Delete(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
