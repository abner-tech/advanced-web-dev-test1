package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependences) routes() http.Handler {
	//setup a new router
	router := httprouter.New()

	//handle 405
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	//method 404
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	//helth check
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthChechHandler)

	//setup routes for the products table d\atabase interaction
	//display all products
	router.HandlerFunc(http.MethodPost, "/v1/products", a.createProductHandler)
	//display a specific product
	router.HandlerFunc(http.MethodGet, "/v1/product/:pid", a.displayProductHandler)
	//update a specific product
	router.HandlerFunc(http.MethodPatch, "/v1/product/:pid", a.updateProductHandler)
	//delette a specific product
	router.HandlerFunc(http.MethodDelete, "/v1/product/:pid", a.deleteProductHandler)
	//display all products--includes sorting, filetering and searching
	router.HandlerFunc(http.MethodGet, "/v1/products", a.listProductHandler)

	//setup routes for the reviews table database interactions
	//create a review for a porduct using product id
	router.HandlerFunc(http.MethodPost, "/v1/reviews/:pid", a.create_P_ReviewHandler)
	//list a specific review for a product
	router.HandlerFunc(http.MethodGet, "/v1/product/:pid/review/:rid", a.listSingleProductReviewHandler)
	//update a specific review for a specific product
	router.HandlerFunc(http.MethodPatch, "/v1/product/:pid/review/:rid", a.updateProductReviewByIDS_Handler)
	//delete a review
	router.HandlerFunc(http.MethodDelete, "/v1/product/:pid/review/:rid", a.deleteReviewByIDS_Handler)
	//display all reviews
	router.HandlerFunc(http.MethodGet, "/v1/reviews", a.listReviewHandler)
	//display a specific review for a specific product
	router.HandlerFunc(http.MethodGet, "/v1/prod/reviews/:pid", a.listReviewHandler)

	return a.recoverPanic(router)
}
