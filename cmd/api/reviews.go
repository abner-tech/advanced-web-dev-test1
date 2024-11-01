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

func (a *applicationDependences) updateProductReviewByIDS_Handler(w http.ResponseWriter, r *http.Request) {
	//getting review with 2 passed in parameters (review id and product id)
	review, err := a.fetchReviewByID(w, r)
	if err != nil {
		//error have already been printed at fetchReviewByID() so we just return
		return
	}

	//just declare info which can be updated by person from the existing review
	var incomingData struct {
		Rating     *int8   `json:"rating"`
		ReviewText *string `json:"review_text"`
	}

	//decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	//verifying which fields have been changed
	if incomingData.Rating != nil {
		review.Rating = *incomingData.Rating
	}
	if incomingData.ReviewText != nil {
		review.ReviewText = *incomingData.ReviewText
	}

	//validate
	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	//continue with update
	err = a.reviewModel.UpdateReview(review)
	if err != nil {
		a.serverErrorResponse(w, r, err)
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

func (a *applicationDependences) deleteReviewByIDS_Handler(w http.ResponseWriter, r *http.Request) {

	//first retrieve the id parameters for record to be deleted
	//pid=product.id and rid=review.id
	pid, err := a.readIDParam(r, "pid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}
	rid, err := a.readIDParam(r, "rid")
	//using one if loop to check error for both pid and rid retrieval
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.reviewModel.DeleteReview(pid, rid)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	//displaying the deleted message
	data := envelope{
		"message": fmt.Sprintf("review with review id: %d and product id: %d deletet sucessfully", rid, pid),
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependences) listReviewHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := a.readIDParam(r, "pid")
	if err != nil || productID < 1 {
		//to avoid return
		productID = 0
	}

	//fields added for pagination and sorting
	var queryParameterData struct {
		ReviewText string
		UserName   string
		data.Filters
	}

	//get query parameters from url
	queryParameter := r.URL.Query()
	queryParameterData.ReviewText = a.getSingleQueryParameter(queryParameter, "review_text", "")
	queryParameterData.UserName = a.getSingleQueryParameter(queryParameter, "user_name", "")
	v := validator.New()

	queryParameterData.Filters.Page = a.getSingleIntigerParameter(queryParameter, "page", 1, v)
	queryParameterData.Filters.PageSize = a.getSingleIntigerParameter(queryParameter, "page_size", 10, v)
	queryParameterData.Filters.Sorting = a.getSingleQueryParameter(queryParameter, "sorting", "id")
	queryParameterData.Filters.SortSafeList = []string{"id", "id", "-id", "-id"}

	//validate pagination filters
	data.ValidateFilters(v, queryParameterData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	//call getAllReviews to retrieev all reviews from the DB
	reviews, metadata, err := a.reviewModel.GetAppReviews(queryParameterData.ReviewText, queryParameterData.UserName, queryParameterData.Filters, productID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
			return
		default:
			a.serverErrorResponse(w, r, err)
			return
		}
	}

	data := envelope{
		"products":  reviews,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
