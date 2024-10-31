package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/abner-tech/Test1/internal/validator"
)

// each name begins with uppercase to make them exportable/ public
type Product struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float32   `json:"price"`
	Category      string    `json:"category"`
	ImageUrl      string    `json:"image_url"`
	AverageRating float32   `json:"average-rating"`
	CreatedAt     time.Time `json:"created_at"`
	Version       int32     `json:"version"`
}

// commentModel that expects a connection pool
type ProductModel struct {
	DB *sql.DB
}

// Insert Row to comments table
// expects a pointer to the actual product content
func (c ProductModel) InsertProduct(product *Product) error {
	//the sql query to be executed against the database table
	query := `
	INSERT INTO products (name, description, price, category, image_url)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, version`

	//the actual values to be passed into $1 and $2
	args := []any{product.Name, product.Description, product.Price, product.Category, product.ImageUrl}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the comments database table. We ask for the the
	// id, created_at, and version to be sent back to us which we will use
	// to update the Comment struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&product.ID,
		&product.CreatedAt,
		&product.Version)
}

func ValidateProduct(v *validator.Validator, product *Product) {
	// Validate Name
	v.Check(product.Name != "", "name", "must be provided")
	v.Check(len(product.Name) <= 50, "name", "must not be more than 50 bytes")

	// Validate Description
	v.Check(product.Description != "", "description", "must be provided")
	v.Check(len(product.Description) <= 100, "description", "must not be more than 100 bytes")

	// Validate Price (ensure it is a positive number)
	v.Check(product.Price > 0, "price", "must be a positive value")

	// Validate Category (ensure it is not empty)
	v.Check(product.Category != "", "category", "must be provided")
	v.Check(len(product.Category) <= 50, "category", "must not be more than 50 bytes")

	// Validate ImageUrl (ensure it is a valid URL format and not empty)
	v.Check(product.ImageUrl != "", "image_url", "must be provided")
	v.Check(len(product.ImageUrl) <= 200, "image_url", "must not be more than 200 bytes")
}

// get a comment from DB based on ID
func (p ProductModel) GetProduct(id int64) (*Product, error) {
	//check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	//the sql query to be excecuted against the database table
	query := `
	SELECT id, name, description, price, category, image_url, average_rating, created_at, version
	FROM products
	WHERE id = $1
	`

	//declare a variable of type Product to hold the returned values
	var product Product

	//set 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Category,
		&product.ImageUrl,
		&product.AverageRating,
		&product.CreatedAt,
		&product.Version,
	)
	//check for errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &product, nil
}

func (p ProductModel) GetAllProducts(category string, name string, description string, filters Fileters) ([]*Product, Metadata, error) {
	query := fmt.Sprintf(`
	SELECT COUNT(*) OVER(), id, name, description, price, category, image_url, average_rating, created_at, version
	FROM products
	WHERE (to_tsvector('simple',category) @@
		plainto_tsquery('simple', $1) OR $1 = '')
	AND (to_tsvector('simple',name) @@
		plainto_tsquery('simple',$2) OR $2 = '')
	AND (to_tsvector('simple', description)@@
		plainto_tsquery('simple', $3) OR $3 = '')
	ORDER BY %s %s, id ASC
	LIMIT $4 OFFSET $5
	`, filters.sortColumn(), filters.sortDirection())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, category, name, description, filters.limit(), filters.offset())
	//check for errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, Metadata{}, err
		default:
			return nil, Metadata{}, err
		}
	}

	defer rows.Close()
	totalRecords := 0
	products := []*Product{}

	for rows.Next() {
		var prod Product
		err := rows.Scan(&totalRecords, &prod.ID, &prod.Name, &prod.Description, &prod.Price, &prod.Category, &prod.ImageUrl, &prod.AverageRating, &prod.CreatedAt, &prod.Version)
		if err != nil {
			return nil, Metadata{}, err
		}
		products = append(products, &prod)
	}
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	//create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return products, metadata, nil
}

// update  a specific record from the comments table
func (p ProductModel) UpdateProducts(product *Product) error {
	//the sql query to be excecuted against the DB table
	//Every time make an update, version number is incremented

	query := `
	UPDATE products
	SET name=$1, description=$2, price=$3, category=$4, image_url=$5, version=version+1
	WHERE id = $6
	RETURNING version
	`

	args := []any{product.Name, product.Description, product.Price, product.Category, product.ImageUrl, product.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return p.DB.QueryRowContext(ctx, query, args...).Scan(&product.Version)

}

// delete a specific comment form the comments table
func (p ProductModel) DeleteProducts(id int64) error {
	//check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}

	//sql querry to be excecuted against the database table
	query := `
	DELETE FROM products
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// ExecContext does not return any rows unlike QueryRowContext.
	// It only returns  information about the the query execution
	// such as how many rows were affected
	result, err := p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	//maybe wrong id for record was given so we sort of try checking
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (p ProductModel) ProductExist(id int64) (int64, error) {
	if id < 1 {
		return 0, ErrRecordNotFound
	}
	query := `
	SELECT id 
	FROM products
	WHERE id = $1
	LIMIT 1
	`

	var Product Product
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, id).Scan(&Product.ID)
	//check for errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrRecordNotFound
		default:
			return 0, err
		}
	}

	print(*&Product.ID)
	return *&Product.ID, nil
}
