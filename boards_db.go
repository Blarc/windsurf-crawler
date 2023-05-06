package main

import (
	"database/sql"
	"sync"
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
}

type BoardsDB struct {
	mu sync.Mutex
	db *sql.DB
}

const file string = "boards.db"

func CreateBoardsDB() (*BoardsDB, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	println("Creating boards table")
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS boards (id INTEGER PRIMARY KEY, post_id INTEGER, title TEXT, price REAL, liters REAL, weight REAL, length REAL, description TEXT, link TEXT, deleted BOOLEAN DEFAULT FALSE)")
	if err != nil {
		return nil, err
	}

	return &BoardsDB{
		db: db,
	}, nil
}

func (db *BoardsDB) GetByPostId(postId int64) (*Board, error) {
	row := db.db.QueryRow("SELECT id, post_id, title, price, liters, weight, length, description, link FROM boards WHERE post_id = ?", postId)

	board := Board{}
	var err error
	if err = row.Scan(&board.Id, &board.PostId, &board.Title, &board.Price, &board.Liters, &board.Weight, &board.Length, &board.Description, &board.Link); err != nil {
		return nil, err
	}
	return &board, err
}

func (db *BoardsDB) Insert(board Board) (int, error) {
	res, err := db.db.Exec("INSERT INTO boards (post_id, title, price, liters, weight, length, description, link) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", board.PostId, board.Title, board.Price, board.Liters, board.Weight, board.Length, board.Description, board.Link)
	if err != nil {
		return 0, err
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return 0, err
	}
	return int(id), nil
}

func (db *BoardsDB) Update(board Board) error {
	// Update the board
	_, err := db.db.Exec("UPDATE boards SET title = ?, price = ?, liters = ?, weight = ?, length = ?, description = ?, link = ?, deleted = ? WHERE post_id = ?", board.Title, board.Price, board.Liters, board.Weight, board.Length, board.Description, board.Link, board.Deleted, board.PostId)
	return err
}

func (db *BoardsDB) SetDeletedAll() error {
	_, err := db.db.Exec("UPDATE boards SET deleted = TRUE")
	return err
}

func (db *BoardsDB) Close() error {
	return db.db.Close()
}
