package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/abner-tech/Test1/internal/data"
	"github.com/abner-tech/Test1/internal/validator"
)

func (a *applicationDependences) createProductHandler(w http.ResponseWriter, r *http.Request) {
	//create a struct to hold a comment
	//we use struct tags [` `] to make the names display in lowercase
	var incomingData struct {
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		Price         float32 `json:"price"`
		Category      string  `json:"category"`
		ImageUrl      string  `json:"image_url"`
		AverageRating float32 `json:"average_rating"`
	}

	//perform decoding

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	product := &data.Product{
		Name:          incomingData.Name,
		Description:   incomingData.Description,
		Price:         incomingData.Price,
		Category:      incomingData.Category,
		ImageUrl:      incomingData.ImageUrl,
		AverageRating: incomingData.AverageRating,
	}

	v := validator.New()
	//do validation
	data.ValidateProduct(v, product)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) //implemented later
		return
	}

	//add product to the products table in database
	err = a.productModel.InsertProduct(product)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	//for now display the result
	// fmt.Fprintf(w, "%+v\n", incomingData)

	//set a location header, the path to the newly created comments
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/product/%d", product.ID))

	//send a json response with a 201 (new reseource created) status code
	data := envelope{
		"product": product,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependences) fetchProductByID(w http.ResponseWriter, r *http.Request) (*data.Product, error) {
	// Get the id from the URL /v1/comments/:id so that we
	// can use it to query the comments table. We will
	// implement the readIDParam() function later
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
	}

	// Call Get() to retrieve the comment with the specified id
	product, err := a.productModel.GetProduct(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}

	}
	return product, nil
}

func (a *applicationDependences) displayProductHandler(w http.ResponseWriter, r *http.Request) {

	product, err := a.fetchProductByID(w, r)
	if err != nil {
		return
	}
	// display the comment
	data := envelope{
		"product": product,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependences) updateProductHandler(w http.ResponseWriter, r *http.Request) {

	product, err := a.fetchProductByID(w, r)
	if err != nil {
		return
	}

	// Use our temporary incomingData struct to hold the data
	// Note: I have changed the types to pointer to differentiate
	// between the client leaving a field empty intentionally
	// and the field not needing to be updated
	var incomingData struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Price       *float32 `json:"price"`
		Category    *string  `json:"category"`
		ImageUrl    *string  `json:"image_url"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData.Name is nil, no update was provided
	if incomingData.Name != nil {
		product.Name = *incomingData.Name
	}
	if incomingData.Description != nil {
		product.Description = *incomingData.Description
	}
	if incomingData.Price != nil {
		product.Price = *incomingData.Price
	}
	if incomingData.Category != nil {
		product.Category = *incomingData.Category
	}
	if incomingData.ImageUrl != nil {
		product.ImageUrl = *incomingData.ImageUrl
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateProduct(v, product)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// perform the update
	err = a.productModel.UpdateProducts(product)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"product": product,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependences) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}
	err = a.productModel.DeleteProducts(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	//diisplay the product
	data := envelope{
		"message": "product deleted successfully",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependences) listProductHandler(w http.ResponseWriter, r *http.Request) {
	//create a struct to hold the query parameters
	//Later, fields will be added for pagination and sorting (filters)
	var queryParameterData struct {
		Category    string
		Name        string
		Description string
		data.Fileters
	}

	//get query parameters from url
	queryParameter := r.URL.Query()

	//load the query parameters into the created struct
	queryParameterData.Category = a.getSingleQueryParameter(queryParameter, "category", "")
	queryParameterData.Name = a.getSingleQueryParameter(queryParameter, "name", "")
	queryParameterData.Description = a.getSingleQueryParameter(queryParameter, "description", "")
	v := validator.New()

	queryParameterData.Fileters.Page = a.getSingleIntigerParameter(queryParameter, "page", 1, v)
	queryParameterData.Fileters.PageSize = a.getSingleIntigerParameter(queryParameter, "page_size", 10, v)
	queryParameterData.Fileters.Sorting = a.getSingleQueryParameter(queryParameter, "sorting", "id")
	queryParameterData.Fileters.SortSafeList = []string{"id", "name", "-id", "-name"}

	//check validity of filters
	data.ValidateFilters(v, queryParameterData.Fileters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	//call GetAll to retrieve all comments of the DB
	products, metadata, err := a.productModel.GetAllProducts(queryParameterData.Category, queryParameterData.Name, queryParameterData.Description, queryParameterData.Fileters)
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
		"products":  products,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
