package main

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	"log"
	"os"
	"sync"
	"time"
)

type Board struct {
	Id          int64
	PostId      int64
	Title       string
	Price       float64
	Liters      float64
	Weight      float64
	Length      float64
	Description string
	Link        string
	Deleted     bool
	Timestamp   time.Time
}

type BoardsDB struct {
	mu sync.Mutex
	db *sql.DB
}

const file string = "boards.db"

func CreateBoardsDB() (*BoardsDB, error) {

	// Read postgres URL
	postgresUrl := os.Getenv("POSTGRES_URL")
	if len(postgresUrl) == 0 {
		return nil, errors.New("environment variable \"POSTGRES_URL\" is not set")
	}

	// Open database connection
	db, err := sql.Open("postgres", postgresUrl)
	if err != nil {
		return nil, err
	}

	// Create boards table
	log.Println("Creating boards table")
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS boards (id SERIAL PRIMARY KEY, post_id INTEGER, title TEXT, price REAL, liters REAL, weight REAL, length REAL, description TEXT, link TEXT, deleted BOOLEAN DEFAULT FALSE, timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return nil, err
	}

	return &BoardsDB{
		db: db,
	}, nil
}

func (db *BoardsDB) GetByPostId(postId int64) (*Board, error) {
	row := db.db.QueryRow(`
		SELECT *
		FROM boards
		WHERE post_id = $1`,
		postId,
	)

	board := Board{}
	var err error
	if err = row.Scan(&board.Id, &board.PostId, &board.Title, &board.Price, &board.Liters, &board.Weight, &board.Length, &board.Description, &board.Link, &board.Deleted, &board.Timestamp); err != nil {
		return nil, err
	}
	return &board, err
}

func (db *BoardsDB) Insert(board Board) (int, error) {
	var id int64
	err := db.db.QueryRow(`
		INSERT INTO boards (post_id, title, price, liters, weight, length, description, link)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id`,
		board.PostId, board.Title, board.Price, board.Liters, board.Weight, board.Length, board.Description, board.Link,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (db *BoardsDB) Update(board Board) error {
	// Update the board
	_, err := db.db.Exec(`
		UPDATE boards
		SET title = $1, price = $2, liters = $3, weight = $4, length = $5, description = $6, link = $7, deleted = $8
	  	WHERE post_id = $9`,
		board.Title, board.Price, board.Liters, board.Weight, board.Length, board.Description, board.Link, board.Deleted, board.PostId,
	)
	return err
}

func (db *BoardsDB) SetDeletedAll() error {
	_, err := db.db.Exec("UPDATE boards SET deleted = TRUE")
	return err
}

func (db *BoardsDB) Close() error {
	return db.db.Close()
}
