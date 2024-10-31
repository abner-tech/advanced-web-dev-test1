package main

import (
	"errors"
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

	//verifying that product review is being made for exists
	id := a.productIdExist(w, r)
	if id <= 0 {
		//error is already printed if no record was found in the productidExists()
		return
	}
	//adding review to the review table in DB
	err = a.reviewModel.InsertReview(review, id)
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

func (a *applicationDependences) fetchReviewByID(w http.ResponseWriter, r *http.Request) (*data.Review, error) {
	pid, err := a.readIDParam(r, "pid")
	if err != nil {
		a.notFoundResponse(w, r)
		return nil, err
	}

	rid, err := a.readIDParam(r, "rid")
	if err != nil {
		a.notFoundResponse(w, r)
		return nil, err
	}

	review, err := a.reviewModel.GetReviewByIDS(rid, pid)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
			return nil, err
		default:
			a.serverErrorResponse(w, r, err)
		}

	}
	return review, nil
}

func (a *applicationDependences) listSingleProductReviewHandler(w http.ResponseWriter, r *http.Request) {
	review, err := a.fetchReviewByID(w, r)
	if err != nil {
		//error was already printed before so we just come out of function
		return
	}

	data := envelope{
		"review": review,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
