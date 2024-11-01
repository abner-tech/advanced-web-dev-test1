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

	//setup routes for the products table database interaction
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthChechHandler)
	router.HandlerFunc(http.MethodPost, "/v1/products", a.createProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/product/:pid", a.displayProductHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/product/:pid", a.updateProductHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/product/:pid", a.deleteProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products", a.listProductHandler)

	//setup routes for the reviews table database interactions
	router.HandlerFunc(http.MethodPost, "/v1/reviews/:pid", a.create_P_ReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/product/:pid/review/:rid", a.listSingleProductReviewHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/product/:pid/review/:rid", a.updateProductReviewByIDS_Handler)
	router.HandlerFunc(http.MethodDelete, "/v1/product/:pid/review/:rid", a.deleteReviewByIDS_Handler)
	router.HandlerFunc(http.MethodGet, "/v1/reviews", a.listReviewHandler)

	return a.recoverPanic(router)
}
