package main

import (
	"context"
	"fmt"
	"log"

	"e-document-backend/internal/app/document"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@127.0.0.1:5433/e_document_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := document.NewPostgresRepository(pool)
	id := uuid.MustParse("d817198a-250b-4df3-8c2b-0d3289837a12")
	
	err = repo.Delete(context.Background(), id)
	if err != nil {
		fmt.Println("ERROR DELETE:", err)
		return
	}
	fmt.Println("SUCCESS DELETE")
}
