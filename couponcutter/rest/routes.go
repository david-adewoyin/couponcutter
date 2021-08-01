package rest

import (
	"context"
	"couponcutter/authetication"
	"couponcutter/listing"
	"couponcutter/storemanagement"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/httplog"
	"github.com/rs/zerolog"
)

//Server starts an http server
type Server struct {
	router   chi.Router
	listing  listing.Service
	sManager storemanagement.Service
	auth     authetication.Service
	*http.Server
}

//NewServer returns a server object
func NewServer(logger zerolog.Logger, l listing.Service, s storemanagement.Service,
	auth authetication.Service, addr string) Server {
	if l == nil || s == nil || auth == nil || addr == "" {
		log.Println("Unable to start server because services were not configured")
		os.Exit(1)
	}

	router := chi.NewRouter()
	server := Server{
		listing:  l,
		sManager: s,
		router:   router,
		auth:     auth,
		Server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
	router.Use(httplog.RequestLogger(logger))
	server.Routes()
	return server
}

//Routes Adds routes to the server router
func (s *Server) Routes() {

	s.router.Post("/user/login", login(s.auth))
	s.router.Post("/user/signup", signUp(s.auth))
	s.router.Get("/coupon/categories", getCategoriesList(s.listing))
	s.router.Get("/search/categories", getSearchCategories(s.listing))

	s.router.Get("/category/{name}/coupons", getCategoryCoupons(s.listing))
	s.router.Get("/coupon/latest", getLatestCoupons(s.listing))
	s.router.Get("/coupon/popular", getPopularCoupons(s.listing))
	s.router.Get("/coupon/{id}", getSingleCoupon(s.listing))

	s.router.Get("/coupon/qrcode/{id}", getQrCode(s.listing))
	s.router.Get("/store/{id}", getStoreDetails(s.listing))
	s.router.Get("/store/{id}/coupons", getStoreCoupons(s.listing))

	s.router.Post("/store/coupon/{id}/verify", verifyCoupon(s.sManager, s.auth))
	s.router.Post("/store/coupon/{id}/checkstate", checkCouponState(s.sManager))

	s.router.Get("/store", getUserStoreDetails(s.sManager, s.auth))
	s.router.Get("/store/dashboard/coupons", getUserStoreCoupons(s.sManager, s.auth))
	s.router.Get("/store/dashboard/coupons/redeemedcount", getUserStoreCouponsRedeemedCount(s.sManager, s.auth))
	s.router.Get("/store/dashboard/followed", getStoresFollowedByUser(s.listing, s.auth))
	s.router.Get("/store/dashboard/coupons/saved", getCouponsSavedByUser(s.listing, s.auth))

	s.router.Post("/store/dashboard/coupon", couponAction(s.sManager, s.auth))
	s.router.Get("/store/dashboard/employee", getEmployees(s.sManager, s.auth))
	s.router.Post("/store/dashboard/employee", employeeAction(s.sManager, s.auth))
	//s.router.Get("/store/dashboard/substore", getSubStores(s.sManager, s.auth))
	//s.router.Post("/store/dashboard/substore", subStoreAction(s.sManager, s.auth))
	s.router.Post("/store/dashboard/store", storeAction(s.sManager, s.auth))

	h := http.FileServer(http.Dir(""))
	s.router.Handle("/static/images/*", h)

}

//fetch the coupon associated with the coupon id
func getSingleCoupon(lister listing.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		rw.Header().Set("Content-Type", "application/json")
		fmt.Println("reust \n\n")
		coupon, err := lister.SingleCoupon(r.Context(), id)
		fmt.Println("curreent here\n\n")
		if err != nil {
			if errors.Is(err, listing.ErrCouponNotFound) {
				// if id is not associated with any coupon
				r := constructError(http.StatusNotFound,
					"bad coupon id",
					"no coupon is  associated with the id sent to the server")

				rw.WriteHeader(http.StatusNotFound)
				json.NewEncoder(rw).Encode(r)
				return

			}
			fmt.Println(err)
			// error to problem while fetching the coupon
			r := constructError(http.StatusInternalServerError,
				"unable to process request",
				"A problem occurs while processing the request")

			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(r)
			return

		}
		//serving coupon associated with the id
		rw.WriteHeader(http.StatusOK)
		err = json.NewEncoder(rw).Encode(coupon)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getSearchCategories(lister listing.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		list, err := lister.Categories(r.Context())
		if err != nil {
			// server problem while answering the request
			r := constructError(http.StatusInternalServerError,
				"unable to process request",
				"A problem occurs while processing the request")

			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(r)
			return
		}

		rw.WriteHeader(http.StatusOK)
		err = json.NewEncoder(rw).Encode(list)
		if err != nil {
			fmt.Println(err)
		}
	}
}

//fetches the list of popular Categories
func getCategoriesList(lister listing.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		var list *listing.CategoriesResponse

		limit := r.FormValue("limit")
		top := r.FormValue("top")

		l, err := strconv.ParseInt(limit, 10, 2)
		if err != nil {
			l = 100
		}
		if top == "true" {
			list, err = lister.TopCategories(r.Context(), int(l))
		} else {
			list, err = lister.Categories(r.Context())
		}

		if err != nil {
			// server problem while answering the request
			r := constructError(http.StatusInternalServerError,
				"unable to process request",
				"A problem occurs while processing the request")

			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(r)
			return
		}

		rw.WriteHeader(http.StatusOK)
		err = json.NewEncoder(rw).Encode(list)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func signUp(auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		email := r.PostFormValue("email")
		password := r.PostFormValue("password")
		token, err := auth.CreateUser(r.Context(), email, password)

		if err != nil {
			if errors.Is(err, authetication.ErrIdentityAlreadyExists) {
				r := constructError(http.StatusUnprocessableEntity,
					"unable to process request",
					"An account is already associated with this email address")

				rw.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(rw).Encode(r)
				return
			}
			if errors.Is(err, authetication.ErrInvalidEmail) {
				r := constructError(http.StatusUnprocessableEntity,
					"bad email format",
					"Email address is not of good format")

				rw.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(rw).Encode(r)
				return
			}
			if errors.Is(err, authetication.ErrPasswordLengthUnAcceptable) {
				r := constructError(http.StatusUnprocessableEntity,
					"unacceptable password length",
					"Password lenght should be between 6 and 64")
				rw.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(rw).Encode(r)
				return
			}

			r := constructError(http.StatusInternalServerError,
				"unable to process request",
				"A problem occurs while processing the request")

			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(r)
			return
		}
		_ = token
		rw.WriteHeader(http.StatusAccepted)
	}
}

//logins in a user with an authorization token
func login(auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		email := r.PostFormValue("email")
		password := r.PostFormValue("password")
		token, err := auth.Login(r.Context(), email, password)

		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			var jsonErr error
			if errors.Is(err, authetication.ErrInvalidEmail) {
				errorRes := constructErrorWithField(http.StatusConflict,
					"email",
					"invalid email address",
					"Enter a proper email address")
				rw.WriteHeader(http.StatusConflict)
				jsonErr = json.NewEncoder(rw).Encode(errorRes)
				return

			} else if errors.Is(err, authetication.ErrPasswordLengthUnAcceptable) {
				errorRes := constructErrorWithField(http.StatusConflict,
					"password",
					"password is unacceptable",
					"Password length is either too short or too long")
				rw.WriteHeader(http.StatusConflict)
				jsonErr = json.NewEncoder(rw).Encode(errorRes)
				return

			} else if errors.Is(err, authetication.ErrIdentityDoesNotExists) {
				errorRes := constructError(http.StatusConflict,
					"identity does not exists",
					"Email or password is not correct",
				)
				rw.WriteHeader(http.StatusConflict)
				jsonErr = json.NewEncoder(rw).Encode(errorRes)
				return

			} else if errors.Is(err, authetication.ErrUnableToProcessRequest) {
				errorRes := constructError(http.StatusInternalServerError,
					"unable to process request",
					"There was problem in processing your request")
				jsonErr = json.NewEncoder(rw).Encode(errorRes)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			if jsonErr != nil {
				log.Println(jsonErr)
			}
			return

		}
		json.NewEncoder(rw).Encode(token)

	}
}
func getLatestCouponOnly(ctx context.Context, lister listing.Service) (*listing.CouponListResponse, *ResponseError) {
	latest, err := lister.LatestCoupons(ctx)

	if err != nil {
		e := constructError(http.StatusInternalServerError,
			"unable to process request",
			"an error occured while processing your request")

		return nil, &e
	}
	return latest, nil

}
func getFilteredCouponOnly(ctx context.Context, lister listing.Service, filter string) (*listing.CouponListResponse, *ResponseError) {
	// filter the coupons
	filtered, err := lister.LatestCouponsWithFiltering(ctx, filter)
	if err != nil {
		log.Println(err)
		e := constructError(http.StatusInternalServerError,
			"unable to process request",
			"an error occured while processing your request")

		return nil, &e

	}

	return filtered, nil

}
func getLatestCoupons(lister listing.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		filter := r.FormValue("filter")
		lastID := r.FormValue("couponid")
		lastTime := r.FormValue("lasttime")

		// check if last coupon id pass into the query to load more coupons
		if lastID == "" || lastTime == "" {
			// check if filter is empty and return the latest coupons
			if filter == "" {
				coupons, resErr := getLatestCouponOnly(r.Context(), lister)
				if resErr != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(rw).Encode(resErr)
					return
				}
				json.NewEncoder(rw).Encode(coupons)
				return
			}

			filtered, resErr := getFilteredCouponOnly(r.Context(), lister, filter)
			if resErr != nil {
				json.NewEncoder(rw).Encode(resErr)
				return
			}

			json.NewEncoder(rw).Encode(filtered)
			return

		}

		before, err := lister.CouponBeforeIDAndTimeFiltered(r.Context(), lastID, lastTime, filter)
		if err != nil {

			resErr := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")

			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(resErr)
			return
		}
		json.NewEncoder(rw).Encode(before)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getCategoryCoupons(lister listing.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		catName := chi.URLParam(r, "name")

		lastID := r.FormValue("couponid")
		lastTime := r.FormValue("lasttime")

		var coupons *listing.CouponListResponse
		var err error
		if lastID == "" || lastTime == "" {
			coupons, err = lister.CategoryCoupons(r.Context(), catName)
		} else {
			coupons, err = lister.CategoryCouponsBeforeIDAndTime(r.Context(), catName, lastID, lastTime)

		}
		if err != nil {
			resErr := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")

			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(resErr)
			return
		}
		json.NewEncoder(rw).Encode(coupons)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getPopularCoupons(lister listing.Service) http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		filter := r.FormValue("filter")
		var popular *listing.CouponListResponse
		var err error
		if filter != "" {
			popular, err = lister.PopularCouponsWithFiltering(r.Context(), filter)
		} else {
			popular, err = lister.PopularCoupons(r.Context())
		}

		if err != nil {
			log.Println(err)

			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return

		}
		json.NewEncoder(rw).Encode(popular)
	}
}

//returns the store details of the store asscoiated with the store id
func getStoreDetails(lister listing.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		storeID := chi.URLParam(r, "id")
		store, err := lister.StoreDetails(r.Context(), storeID)

		if err != nil {
			fmt.Println(err)
			if errors.Is(err, listing.ErrStoreNotFound) {
				e := constructError(http.StatusNotFound, "content not found", "store with id is not found")
				rw.WriteHeader(http.StatusNotFound)
				json.NewEncoder(rw).Encode(e)
				return
			}

			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return
		}
		err = json.NewEncoder(rw).Encode(store)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// returns the  coupons associated with a particular store
func getStoreCoupons(lister listing.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")
		storeID := chi.URLParam(r, "id")
		lastID := r.FormValue("last")
		lastTime := r.FormValue("time")

		var coupons *listing.CouponListResponse
		var err error
		if lastID != "" && lastTime != "" {
			coupons, err = lister.StoreCouponsBeforeIDAndTime(r.Context(), storeID, lastID, lastTime)
		} else {
			coupons, err = lister.StoreCoupons(r.Context(), storeID)
		}
		if err != nil {
			if err != nil {
				if errors.Is(err, listing.ErrStoreNotFound) {
					e := constructError(http.StatusNotFound, "content not found", "store with id is not found")
					rw.WriteHeader(http.StatusNotFound)
					json.NewEncoder(rw).Encode(e)
					return
				}

				e := constructError(http.StatusInternalServerError,
					"unable to process request",
					"an error occured while processing your request")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(e)
				return
			}
		}
		json.NewEncoder(rw).Encode(coupons)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
}

func getUserStoreDetails(sManager storemanagement.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "no token found", "user authetication is required for this route")
			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		userid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		userStore, err := sManager.UserStoreData(r.Context(), userid)

		if err != nil {
			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return
		}
		err = json.NewEncoder(rw).Encode(userStore)
		if err != nil {

			fmt.Println(err)
		}
	}
}

// returns a list of coupons of the user store
func getUserStoreCoupons(sManager storemanagement.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")

		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "no token found", "user authetication is required for this route")
			fmt.Println(e)
			rw.WriteHeader(http.StatusUnauthorized)

			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			fmt.Println("here no w")

			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		userid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		coupons, err := sManager.UserStoreCoupons(r.Context(), userid)
		fmt.Println("here")

		if err != nil {
			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return
		}
		err = json.NewEncoder(rw).Encode(coupons)
		if err != nil {

			fmt.Println(err)
		}
	}
}

// returns a list of coupons of the user store
func getUserStoreCouponsRedeemedCount(sManager storemanagement.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")
		filter := r.FormValue("filter")

		authBearer := r.Header.Get("authorization")

		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "no token found", "user authetication is required for this route")
			fmt.Println(e)
			rw.WriteHeader(http.StatusUnauthorized)

			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			fmt.Println("here no w")

			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		userid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		count, err := sManager.GetUserStoreCouponsRedeemedCount(r.Context(), userid, filter)

		if err != nil {
			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return
		}
		type Response struct {
			Count uint `json:"count,omitempty"`
		}
		err = json.NewEncoder(rw).Encode(Response{Count: count})
		if err != nil {
			fmt.Println(err)
		}
	}
}

// returns the list of coupons saved by a particular user
func getCouponsSavedByUser(lister listing.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")
		lastID := r.FormValue("last")
		lastTime := r.FormValue("time")

		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "no token found", "user authetication is required for this route")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		userid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		var coupons *listing.CouponListResponse
		var err error

		if lastID == "" || lastTime == "" {
			coupons, err = lister.CouponsSavedByUser(r.Context(), userid)
		} else {
			coupons, err = lister.CouponBeforeIDAndTimeFiltered(r.Context(), userid, lastID, lastTime)
		}
		if err != nil {
			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return

		}
		rw.WriteHeader(http.StatusOK)
		err = json.NewEncoder(rw).Encode(coupons)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
}

// returns the list of stores followed by a particular user
func getStoresFollowedByUser(lister listing.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")

		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "no token found", "user authetication is required for this route")
			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		userid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		stores, err := lister.StoresFollowedByUser(r.Context(), userid)

		if err != nil {
			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return
		}
		err = json.NewEncoder(rw).Encode(stores)
		if err != nil {

			fmt.Println(err)
		}
	}
}

// perform certain actions on employees management
func employeeAction(sManager storemanagement.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")
		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		userid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		type EmployeePayload struct {
			EmpID  string `json:"emp_id,omitempty"`
			Email  string `json:"email,omitempty"`
			Action string `json:"action,omitempty"`
		}

		if r.Body == nil {
			e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
			rw.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(rw).Encode(e)
			return
		}

		payload := &EmployeePayload{}
		err := json.NewDecoder(r.Body).Decode(payload)
		if err != nil {
			e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
			rw.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(rw).Encode(e)
			return
		}

		if payload.Action == "suspend" || payload.Action == "remove" || payload.Action == "resume" {
			if payload.EmpID == "" {
				e := constructError(http.StatusUnprocessableEntity, "no valid body found", "retry the request by sending a body")
				rw.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(rw).Encode(e)
				return
			}
		}

		if payload.Email == "" {
			if payload.Action == "create" {

				e := constructError(http.StatusUnprocessableEntity, "no valid body found", "retry the request by sending a body")
				rw.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(rw).Encode(e)
				return
			}
			err := sManager.AddEmployee(r.Context(), userid, storemanagement.Employee{
				Email: payload.Email,
			})
			if err != nil {
				e := constructError(http.StatusInternalServerError,
					"unable to process request",
					"an error occured while processing your request")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(e)
				return
			}
			type ActionResponse struct {
				Type  string `json:"type"`
				Email string `json:"email"`
				State bool   `json:"create"`
			}
			rw.WriteHeader(http.StatusAccepted)
			json.NewEncoder(rw).Encode(ActionResponse{
				Type:  "employee",
				Email: payload.Email,
				State: true})
			return
		}

		if payload.Action == "suspend" {
			fmt.Printf("suspend")
			err := sManager.SuspendEmployee(r.Context(), userid, payload.EmpID)
			if err != nil {
				if errors.Is(err, storemanagement.ErrEmployeeNotFound) {
					e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
					rw.WriteHeader(http.StatusUnprocessableEntity)
					json.NewEncoder(rw).Encode(e)
					return
				}
				e := constructError(http.StatusInternalServerError,
					"unable to process request",
					"an error occured while processing your request")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(e)
				return
			}
			type ActionResponse struct {
				Type  string `json:"type,omitempty"`
				EmpID string `json:"emp_id,omitempty"`
				State bool   `json:"suspended,omitempty"`
			}
			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(ActionResponse{
				Type:  "employee",
				EmpID: payload.EmpID,
				State: true})
			return
		}
		if payload.Action == "remove" {
			fmt.Printf("removing")
			err := sManager.RemoveEmployee(r.Context(), userid, payload.EmpID)
			if err != nil {
				if errors.Is(err, storemanagement.ErrEmployeeNotFound) {
					e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
					rw.WriteHeader(http.StatusUnprocessableEntity)
					json.NewEncoder(rw).Encode(e)
					return
				}
				e := constructError(http.StatusInternalServerError,
					"unable to process request",
					"an error occured while processing your request")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(e)
				return
			}
			type ActionResponse struct {
				Type  string `json:"type,omitempty"`
				EmpID string `json:"emp_id,omitempty"`
				State bool   `json:"removed,omitempty"`
			}
			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(ActionResponse{
				Type:  "employee",
				EmpID: payload.EmpID,
				State: true})
			return

		}
		if payload.Action == "resume" {
			fmt.Printf("resuming")
			err := sManager.ResumeEmployee(r.Context(), userid, payload.EmpID)
			if err != nil {
				if errors.Is(err, storemanagement.ErrEmployeeNotFound) {
					e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
					rw.WriteHeader(http.StatusUnprocessableEntity)
					json.NewEncoder(rw).Encode(e)
					return
				}
				e := constructError(http.StatusInternalServerError,
					"unable to process request",
					"an error occured while processing your request")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(e)
				return
			}
			type ActionResponse struct {
				Type  string `json:"type,omitempty"`
				EmpID string `json:"emp_id,omitempty"`
				State bool   `json:"resumed,omitempty"`
			}
			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(ActionResponse{
				Type:  "employee",
				EmpID: payload.EmpID,
				State: true,
			})
			return

		}

		// if action is not found
		e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
		rw.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(rw).Encode(e)
		return

	}
}

// perform certain actions on coupons management
func couponAction(sManager storemanagement.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")
		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return

		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		userid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		type CouponPayload struct {
			CouponID      string   `json:"coupon_id,omitempty"`
			AmountOff     float64  `json:"amount_off,omitempty"`
			PercentageOff float64  `json:"percentage_off,omitempty"`
			Desc          string   `json:"desc,omitempty"`
			Categories    []string `json:"Categories"`
			CurrencyCode  string   `json:"currency_code,omitempty"`
			DiscountType  string   `json:"discount_type,omitempty"`
			ExpiringDate  uint     `json:"expiring_date"`
			//	SingleUserUse       bool     `json:"single_user_use,omitempty"`
			UnlimitedRedemption bool   `json:"unlimited_redemption,omitempty"`
			MaxRedemption       uint   `json:"max_redemption,omitempty"`
			Action              string `json:"action,omitempty"`
			IsTextCoupon        bool   `json:"is_text_coupon,omitempty"`
			TextCouponWebURL    string `json:"text_coupon_web_url,omitempty"`
			TextCouponCode      string `json:"text_coupon_code,omitempty"`
		}

		payload := &CouponPayload{}
		err := json.NewDecoder(r.Body).Decode(payload)
		if err != nil {
			fmt.Println(err)
			e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
			rw.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(rw).Encode(e)
			return
		}
		if payload.Action == "create" {
			coupon := storemanagement.CreateCoupon{
				StoreID: userid,

				Desc:         payload.Desc,
				Categories:   payload.Categories,
				CurrencyCode: payload.CurrencyCode,
				ExpiringDate: payload.ExpiringDate,
				//	SingleUserUse:       payload.SingleUserUse,
				UnlimitedRedemption: payload.UnlimitedRedemption,
				MaxRedemption:       payload.MaxRedemption,
				IsTextCoupon:        payload.IsTextCoupon,
				TextCouponCode:      payload.TextCouponCode,
				TextCouponWebURL:    payload.TextCouponWebURL,
				DiscountType:        payload.DiscountType,
			}
			couponid, err := sManager.CreateCoupon(r.Context(), userid, coupon)

			if err != nil {
				fmt.Println(err)
				e := constructError(http.StatusInternalServerError,
					"unable to process request",
					"an error occured while processing your request")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(e)
				return
			}
			type ActionResponse struct {
				Type     string `json:"type"`
				CouponID string `json:"coupon_id"`
				State    bool   `json:"created"`
			}
			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(ActionResponse{
				Type:     "coupon",
				CouponID: couponid,
				State:    true,
			})
			return
		}
		if payload.Action == "delete" {
			fmt.Println("here now")
			if payload.CouponID == "" {
				e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
				rw.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(rw).Encode(e)
				return
			}

			sManager.DeleteCoupon(r.Context(), userid, payload.CouponID)
			type ActionResponse struct {
				Type     string `json:"type"`
				CouponID string `json:"coupon_id"`
				State    bool   `json:"deleted"`
			}
			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(ActionResponse{
				Type:     "coupon",
				CouponID: payload.CouponID,
				State:    true,
			})
			return

		}

		// if action is not found
		e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
		rw.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(rw).Encode(e)
		return

	}
}

// perform certain actions on sstore
func storeAction(sManager storemanagement.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")
		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		userid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		type StorePayload struct {
			Name    string `json:"name,omitempty"`
			Tagline string `json:"tagline,omitempty"`
			Address string `json:"address,omitempty"`
			Action  string `json:"action,omitempty"`
		}

		if r.Body == nil {
			e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
			rw.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(rw).Encode(e)
			return
		}

		payload := &StorePayload{}

		err := json.NewDecoder(r.Body).Decode(payload)
		if err != nil {
			e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
			rw.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(rw).Encode(e)
			return
		}

		if payload.Action == "edit" {
			err := sManager.EditStore(r.Context(), userid, storemanagement.StoreEdit{
				Name:    payload.Name,
				Tagline: payload.Tagline,
				Address: payload.Address,
			})

			if err != nil {
				e := constructError(http.StatusInternalServerError,
					"unable to process request",
					"an error occured while processing your request")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(e)
				return
			}
			return
		}
	}
}

/*
// perform certain actions on substore management
func subStoreAction(sManager storemanagement.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")
		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		userid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		type SubStorePayload struct {
			SubStoreID    string `json:"sub_store_id,omitempty"`
			SubStoreEmail string `json:"sub_store_email,omitempty"`
			Action        string `json:"action,omitempty"`
		}

		if r.Body == nil {
			e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
			rw.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(rw).Encode(e)
			return
		}

		payload := &SubStorePayload{}

		err := json.NewDecoder(r.Body).Decode(payload)
		if err != nil {
			e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
			rw.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(rw).Encode(e)
			return
		}

		if payload.Action == "add" {
			if payload.SubStoreEmail == "" {
				e := constructError(http.StatusUnprocessableEntity, "no valid body found", "retry the request by sending a valid body")
				rw.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(rw).Encode(e)
				return
			}
			err := sManager.AddSubStore(userid, payload.SubStoreEmail)
			if err != nil {
				e := constructError(http.StatusInternalServerError,
					"unable to process request",
					"an error occured while processing your request")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(e)
				return
			}

			type ActionResponse struct {
				Type  string `json:"sub_store"`
				Email string `json:"email"`
				State bool   `json:"adding"`
			}
			rw.WriteHeader(http.StatusAccepted)
			json.NewEncoder(rw).Encode(ActionResponse{
				Type:  "sub_store",
				Email: payload.SubStoreEmail,
				State: true})
			return
		}

		if payload.Action == "remove" {
			if payload.SubStoreID == "" {
				e := constructError(http.StatusUnprocessableEntity, "no valid body found", "retry the request by sending a valid body")
				rw.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(rw).Encode(e)
				return
			}
			err := sManager.RemoveSubStore(userid, payload.SubStoreID)
			if err != nil {
				if errors.Is(err, storemanagement.ErrStoreNotFound) {
					e := constructError(http.StatusUnprocessableEntity, "no store found", "no sub store is associated with this id")
					rw.WriteHeader(http.StatusUnprocessableEntity)
					json.NewEncoder(rw).Encode(e)
					return
				}
				e := constructError(http.StatusInternalServerError,
					"unable to process request",
					"an error occured while processing your request")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(e)
				return
			}
			type ActionResponse struct {
				Type       string `json:"type"`
				SubStoreID string `json:"sub_store_id"`
				State      bool   `json:"removed"`
			}
			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(ActionResponse{
				Type:       "substore",
				SubStoreID: payload.SubStoreID, State: true,
			})
			return

		}

		// if action is not found
		e := constructError(http.StatusUnprocessableEntity, "no body found", "retry the request by sending a body")
		rw.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(rw).Encode(e)
		return

	}
} */

//v2
/* // fetch substores associated with store
func getSubStores(sManager storemanagement.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")
		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		storeid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		stores, err := sManager.SubStores(storeid)
		if err != nil {
			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return
		}
		json.NewEncoder(rw).Encode(stores)

	}
}
*/
// fetch employees associated with store
func getEmployees(sManager storemanagement.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("Content-Type", "application/json")
		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		storeid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		employees, err := sManager.Employees(r.Context(), storeid)
		fmt.Printf(" \n\n\n")
		fmt.Println(employees)
		if err != nil {
			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return
		}
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(employees)

	}
}

func getQrCode(lister listing.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		id := chi.URLParam(r, "id")
		qrbytes, err := lister.QrImage(r.Context(), id)
		if err != nil {
			fmt.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("sending qr code")
		_, err = rw.Write(qrbytes)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// checkCouponState checks if a coupon is valid
func checkCouponState(sManager storemanagement.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		couponid := chi.URLParam(r, "id")
		state, err := sManager.CouponState(r.Context(), couponid)
		if err != nil {
			if errors.Is(err, storemanagement.ErrCouponNotValid) {
				e := constructError(http.StatusBadRequest,
					"couponid is not valid",
					"the couponid is not associated with any coupon")
				rw.WriteHeader(e.Code)
				json.NewEncoder(rw).Encode(e)
				return
			}

			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return
		}

		type Action struct {
			CouponID string `json:"coupon_id,omitempty"`
			State    string `json:"state,omitempty"`
		}
		json.NewEncoder(rw).Encode(Action{CouponID: couponid, State: state})

	}
}

// verifyCoupon validate coupons
func verifyCoupon(sManager storemanagement.Service, auth authetication.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		couponid := chi.URLParam(r, "id")

		authBearer := r.Header.Get("authorization")
		if authBearer == "" {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		token := strings.Split(authBearer, " ")
		if len(token) != 2 {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}
		userid, valid := auth.VerifyToken(token[1])
		if !valid {
			e := constructError(http.StatusUnauthorized, "valid token not found", "authetication is needed to make this action")
			rw.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(rw).Encode(e)
			return
		}

		err := sManager.VerifyCoupon(r.Context(), userid, couponid)
		if err != nil {
			if errors.Is(err, storemanagement.ErrCouponIsUsed) {
				e := constructError(http.StatusConflict,
					"coupon is already used",
					"coupon is already used")
				rw.WriteHeader(e.Code)
				json.NewEncoder(rw).Encode(e)
				return
			}
			if errors.Is(err, storemanagement.ErrCouponNotValid) {
				e := constructError(http.StatusBadRequest,
					"couponid is not valid",
					"the couponid is not associated with any coupon")
				rw.WriteHeader(e.Code)
				json.NewEncoder(rw).Encode(e)
				return
			}
			if errors.Is(err, storemanagement.ErrCouponLimitExceeded) {
				e := constructError(http.StatusConflict,
					"coupon limit exceeded",
					"coupon redemption limit is already exceeded")
				rw.WriteHeader(e.Code)
				json.NewEncoder(rw).Encode(e)
				return
			}

			e := constructError(http.StatusInternalServerError,
				"unable to process request",
				"an error occured while processing your request")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(e)
			return
		}
		rw.WriteHeader(http.StatusOK)

	}
}
