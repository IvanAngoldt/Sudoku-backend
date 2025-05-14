package handlers

import (
	"users/database"

	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	db     *database.Database
	logger *logrus.Logger
}

func NewUserHandler(db *database.Database, logger *logrus.Logger) *UserHandler {
	return &UserHandler{db: db, logger: logger}
}
