package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (r ReviewModel) GetReviewByIDS(rid int64, pid int64) (*Review, error) {
	//validate id
	if pid < 1 || rid < 1 {
		return nil, ErrRecordNotFound
	}

	//query
	query := `SELECT id, product_id, user_name, rating, review_text, helpful_count, created_at, version
	FROM reviews
	WHERE id = $1 AND product_id = $2
	`
	var review Review

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, rid, pid).Scan(
		&review.ID,
		&review.ProductID,
		&review.UserName,
		&review.Rating,
		&review.ReviewText,
		&review.HelpfulCount,
		&review.CreatedAt,
		&review.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &review, nil
}

func (r ReviewModel) UpdateReview(review *Review) error {
	query := `
	UPDATE reviews
	SET rating = $1, review_text=$2, version=version+1
	WHERE id=$3 AND product_id=$4
	RETURNING version
	`
	args := []any{review.Rating, review.ReviewText, review.ID, review.ProductID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return r.DB.QueryRowContext(ctx, query, args...).Scan(&review.Version)
}

func (r ReviewModel) DeleteReview(pid int64, rid int64) error {
	//checking if both id's are valid
	if pid < 1 || rid < 1 {
		return ErrRecordNotFound
	}

	//sql querry to delete record from reviews table
	query := `
	DELETE FROM reviews
	WHERE id = $1 AND product_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := r.DB.ExecContext(ctx, query, rid, pid)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (r ReviewModel) GetAppReviews(reviewText string, name string, filters Filters, productID int64) ([]*Review, Metadata, error) {
	// Base query with placeholders for reviewText and name filtering
	query := fmt.Sprintf(`
	SELECT COUNT(*) OVER(), id, product_id, user_name, rating, review_text, helpful_count, created_at, version
	FROM reviews
	WHERE (to_tsvector('simple', review_text) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (to_tsvector('simple', user_name) @@ plainto_tsquery('simple', $2) OR $2 = '')
	`)

	// Add an additional condition if productID is non-zero
	args := []interface{}{reviewText, name}
	if productID != 0 {
		query += "AND product_id = $3 "
		args = append(args, productID)
	}

	// Add ordering and pagination
	query += fmt.Sprintf("ORDER BY %s %s, id ASC LIMIT $%d OFFSET $%d",
		filters.sortColumn(),
		filters.sortDirection(),
		len(args)+1, // LIMIT placeholder index
		len(args)+2, // OFFSET placeholder index
	)
	args = append(args, filters.limit(), filters.offset())

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute query
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	reviews := []*Review{}

	for rows.Next() {
		var rev Review
		err := rows.Scan(&totalRecords,
			&rev.ID,
			&rev.ProductID,
			&rev.UserName,
			&rev.Rating,
			&rev.ReviewText,
			&rev.HelpfulCount,
			&rev.CreatedAt,
			&rev.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		reviews = append(reviews, &rev)
	}
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return reviews, metadata, nil
}
