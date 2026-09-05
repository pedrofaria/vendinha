package main

import "database/sql"

// repo agrupa o acesso a dados (o banco é aberto em NewRepo).
type repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *repo {
	return &repo{db: db}
}
