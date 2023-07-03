package test

import (
	"database/sql"
	"regexp"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type v2Suite struct {
	db      *gorm.DB
	mock    sqlmock.Sqlmock
	student Student
}

type Student struct {
	ID   string
	Name string
}

func TestGORMV2(t *testing.T) {
	s := &v2Suite{}
	var (
		db  *sql.DB
		err error
	)

	db, s.mock, err = sqlmock.New()
	if err != nil {
		t.Errorf("Failed to open mock sql db, got error: %v", err)
	}

	if db == nil {
		t.Error("mock db is null")
	}

	if s.mock == nil {
		t.Error("sqlmock is null")
	}

	s.db, err = gorm.Open(postgres.New(
		postgres.Config{
			Conn:       db,
			DriverName: "postgres",
		},
	), &gorm.Config{})
	if err != nil {
		panic(err) // Error here
	}

	defer db.Close()

	s.student = Student{
		ID:   "123456",
		Name: "Test 1",
	}

	defer db.Close()

	studentID, _ := strconv.Atoi(s.student.ID)

	s.mock.ExpectBegin()

	s.mock.ExpectExec(
		regexp.QuoteMeta(`INSERT INTO "students" ("id","name") VALUES ($1,$2)`)).
		WithArgs(s.student.ID, s.student.Name).
		WillReturnResult(sqlmock.NewResult(int64(studentID), 1))

	s.mock.ExpectCommit()

	if err = s.db.Create(&s.student).Error; err != nil {
		t.Errorf("Failed to insert to gorm db, got error: %v", err)
		t.FailNow()
	}

	err = s.mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet expectations, got error: %v", err)
	}
}
