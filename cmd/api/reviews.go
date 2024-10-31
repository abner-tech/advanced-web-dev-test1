package main

import (
	"fmt"
	"net/http"

	"github.com/abner-tech/Test1/internal/data"
	"github.com/abner-tech/Test1/internal/validator"
)

func (a *applicationDependences) create_P_ReviewHandler(w http.ResponseWriter, r *http.Request) {
	//struct to hold a review
	var incommingData struct {
		ProductID    int64  `json:"product_id"`
		UserName     string `json:"user_name"`
		Rating       int8   `json:"rating"`
		ReviewText   string `json:"review_text"`
		HelpfulCount int8   `json:"helpful_count"`
	}

	//decoding
	err := a.readJSON(w, r, &incommingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	review := &data.Review{
		ProductID:    incommingData.ProductID,
		UserName:     incommingData.UserName,
		Rating:       incommingData.Rating,
		ReviewText:   incommingData.ReviewText,
		HelpfulCount: incommingData.HelpfulCount,
	}

	v := validator.New()
	//implementing validation
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	//verifying that  product exists
	product, err := a.fetchProductByID(w, r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	println(product.Category)

	ID := product.ID

	//adding review to the review table in DB
	err = a.reviewModel.InsertReview(review, ID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	//set location header, path to the newly created review
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/reviews/%d", review.ID))

	data := envelope{
		"review": review,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

// func (a *applicationDependences) fetchReviewByID(w http.ResponseWriter, r *http.Request) (*data.Review, error) {
// 	id, err := a.readIDParam(r)
// 	if err != nil {
// 		a.notFoundResponse(w, r)
// 	}

// 	review, err := a.reviewModel.GetReview(id)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrRecordNotFound):
// 			a.notFoundResponse(w, r)
// 		default:
// 			a.serverErrorResponse(w, r, err)
// 		}

// 	}
// 	return review, nil
// }
