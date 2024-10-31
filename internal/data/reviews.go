package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/abner-tech/Test1/internal/validator"
)

type Review struct {
	ID           int64     `json:"id"`
	ProductID    int64     `json:"product_id"`
	UserName     string    `json:"user_name"`
	Rating       int8      `json:"rating"`
	ReviewText   string    `json:"review_text"`
	HelpfulCount int8      `json:"helpful_count"`
	CreatedAt    time.Time `json:"created_at"`
	Version      int16     `json:"version"`
}

type ReviewModel struct {
	DB *sql.DB
}

func (r ReviewModel) InsertReview(review *Review, ID int64) error {
	//query to excecute
	query := `
	INSERT INTO reviews (product_id, user_name, rating, review_text)
	VALUES($1, $2, $3, $4)
	RETURNING id, product_id, created_at, version
	`

	review.ProductID = ID

	args := []any{review.ProductID, review.UserName, review.Rating, review.ReviewText}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(
		&review.ID,
		&review.ProductID,
		&review.CreatedAt,
		&review.Version,
	)
}

func ValidateReview(v *validator.Validator, review *Review) {
	//validate values
	v.Check(review.UserName != "", "user_name", "must be provided")
	v.Check(len(review.UserName) <= 25, "username", "must not be more than 25 bytes")

	v.Check(review.Rating >= 0 && review.Rating <= 5, "rating", "must be a number between 1 and 5")

	v.Check(review.ReviewText != "", "review_text", "must be provided")
	v.Check(len(review.ReviewText) <= 100, "review_text", "must not be more than 100 bytes")
}

// func (r ReviewModel) GetReview(id int64) (*Review, error) {
// 	if id < 1 {
// 		return nil, ErrRecordNotFound
// 	}

// 	//query
// 	query := `SELECT id
// 	FROM products
// 	WHERE id = $1
// 	`

// }
