package login

import (
	"database/sql"
	"github.com/google/uuid"
	"log/slog"
)

type Login struct {
	db *sql.DB
}

type Uuid struct {
	IdUuid int
	Uuid   string
}

func New(db *sql.DB) *Login {
	return &Login{db: db}
}

func (l *Login) UuidExist(uuid string) bool {
	id := Uuid{}
	slog.Info("check if uuid exists", "query", "SELECT * FROM uuid WHERE uuid=?")
	isExist := l.db.QueryRow("SELECT * FROM uuid WHERE uuid=?", uuid)
	err := isExist.Scan(&id.IdUuid, &id.Uuid)
	if err != nil {
		slog.Info("uuid is not exist")
		return false
	}
	slog.Info("uuid is exist")
	return true
}

func (l *Login) NewId(id string) (string, error) {
	isAuth := l.UuidExist(id)
	if !isAuth {
		slog.Info("the entered uuid does not exist, create a new one")
		newId, _ := uuid.NewRandom()
		slog.Info("add a new uuid to the table", "uuid", newId)
		_, err := l.db.Exec("INSERT INTO uuid(uuid) VALUES (?)", newId.String())
		if err != nil {
			return "", err
		}
		slog.Info("new uuid successfully added")
		return newId.String(), nil
	}
	return "", nil
}
