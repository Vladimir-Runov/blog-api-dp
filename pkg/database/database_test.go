package database

import (
 "regexp"
 "testing"

 "github.com/DATA-DOG/go-sqlmock"
)

func TestGetDSN(t *testing.T) {
 cfg := Config{
  Host:     "localhost",
  Port:     5432,
  User:     "postgres",
  Password: "secret",
  DBName:   "test_db",
  SSLMode:  "disable",
 }

 expected := "host=localhost port=5432 user=postgres password=secret dbname=test_db sslmode=disable"

 actual := GetDSN(cfg)

 if actual != expected {
  t.Errorf("GetDSN() = %q, ожидалось %q", actual, expected)
 }
}

func TestCheckConnection_Success(t *testing.T) {
 db, mock, err := sqlmock.New(
  sqlmock.MonitorPingsOption(true),
 )
 if err != nil {
  t.Fatalf("не удалось создать mock database: %v", err)
 }
 defer db.Close()

 mock.ExpectPing().WillReturnError(nil)

 err = CheckConnection(db)
 if err != nil {
  t.Fatalf("CheckConnection() вернул ошибку: %v", err)
 }

 if err := mock.ExpectationsWereMet(); err != nil {
  t.Fatalf("ожидания mock не выполнены: %v", err)
 }
}

func TestCheckConnection_Error(t *testing.T) {
 db, mock, err := sqlmock.New(
  sqlmock.MonitorPingsOption(true),
 )
 if err != nil {
  t.Fatalf("не удалось создать mock database: %v", err)
 }
 defer db.Close()

 expectedErr := "connection refused"

 mock.ExpectPing().WillReturnError(
  &testError{message: expectedErr},
 )

 err = CheckConnection(db)
 if err == nil {
  t.Fatal("CheckConnection() ожидалась ошибка, но её нет")
 }

 expectedMessage := "нет соединения с базой данных: " + expectedErr
 if err.Error() != expectedMessage {
  t.Errorf(
   "ошибка = %q, ожидалось %q",
   err.Error(),
   expectedMessage,
  )
 }

 if err := mock.ExpectationsWereMet(); err != nil {
  t.Fatalf("ожидания mock не выполнены: %v", err)
 }
}

func TestTestConnection_Success(t *testing.T) {
 db, mock, err := sqlmock.New()
 if err != nil {
  t.Fatalf("не удалось создать mock database: %v", err)
 }
 defer db.Close()

 mock.ExpectQuery(regexp.QuoteMeta("SELECT 1")).
  WillReturnRows(
   sqlmock.NewRows([]string{"result"}).
    AddRow(1),
  )

 err = TestConnection(db)
 if err != nil {
  t.Fatalf("TestConnection() вернул ошибку: %v", err)
 }

 if err := mock.ExpectationsWereMet(); err != nil {
  t.Fatalf("ожидания mock не выполнены: %v", err)
 }
}

func TestTestConnection_Error(t *testing.T) {
 db, mock, err := sqlmock.New()
 if err != nil {
  t.Fatalf("не удалось создать mock database: %v", err)
 }
 defer db.Close()

 mock.ExpectQuery(regexp.QuoteMeta("SELECT 1")).
  WillReturnError(&testError{message: "query failed"})

 err = TestConnection(db)
 if err == nil {
  t.Fatal("TestConnection() ожидалась ошибка, но её нет")
 }

 expectedMessage := "тестовая команда не выполнена: query failed"
 if err.Error() != expectedMessage {
  t.Errorf(
   "ошибка = %q, ожидалось %q",
   err.Error(),
   expectedMessage,
  )
 }

 if err := mock.ExpectationsWereMet(); err != nil {
  t.Fatalf("ожидания mock не выполнены: %v", err)
 }
}

func TestClose_Success(t *testing.T) {
 db, mock, err := sqlmock.New()
 if err != nil {
  t.Fatalf("не удалось создать mock database: %v", err)
 }

 mock.ExpectClose()

 err = Close(db)
 if err != nil {
  t.Fatalf("Close() вернул ошибку: %v", err)
 }

 if err := mock.ExpectationsWereMet(); err != nil {
  t.Fatalf("ожидания mock не выполнены: %v", err)
 }
}

// testError нужен для имитации ошибок базы данных.
type testError struct {
 message string
}

func (e *testError) Error() string {
 return e.message
}
