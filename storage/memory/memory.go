package memory

import (
	"couponcutter/listing"
	"couponcutter/storemanagement"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"os"

	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

//MaxRedemption represent the limit the maximum  number a coupon should be redeem
type MaxRedemption string

//COST for the hashing
const COST = 14

// Coupon represent a redeemable piece of text or qrcode
type Coupon struct {
	CouponID       string        `json:"coupon_id,omitempty"`
	Store          Store         `json:"store,omitempty"`
	AmountOff      float64       `json:"AmountOff,omitempty"`
	PercentageOff  float64       `json:"percentageOff,omitempty"`
	Desc           string        `json:"desc,omitempty"`
	CurrencyCode   string        `json:"currency,omitempty"`
	CreationDate   uint          `json:"creation_date,omitempty"`
	ExpiringDate   uint          `json:"expiring_date,omitempty"`
	Categories     []string      `json:"categories,omitempty"`
	ThemeColor     int           `json:"theme_color,omitempty"`
	MaxRedemption  MaxRedemption `json:"max_redemption,omitempty"`
	SingleUserUse  bool          `json:"single_use,omitempty"`
	IsTextCoupon   bool          `json:"text_only,omitempty"`
	TextCouponCode string        `json:"text_coupon_code,omitempty"`
	QrCodeURL      string        `json:"item_url,omitempty"`
	DiscountType   string        `json:"discount_type,omitempty"`
	// voucher percentage
}

//Employee is an identifiable entity that has limited store_management abilities
type Employee struct {
	FullName string `json:"full_name,omitempty"`
	Email    string `json:"email,omitempty"`
	State    string `json:"state,omitempty"`
}

// Store represent an indentity that owns coupons
type Store struct {
	StoreName   string     `json:"store_name,omitempty"`
	StoreID     string     `json:"store_id,omitempty"`
	StoreImage  string     `json:"StoreImage,omitempty"`
	Tagline     string     `json:"tagline,omitempty"`
	Address     string     `json:"address,omitempty"`
	Following   bool       `json:"following,omitempty"`
	CouponCount uint       `json:"coupon_count,omitempty"`
	Employees   []Employee `json:"employees,omitempty"`
	SubStores   []Store    `json:"sub_stores,omitempty"`
	Coupons     []Coupon   `json:"coupons,omitempty"`
	ThemeColor  int        `json:"theme_color,omitempty"`
}

//Storage provides access to a storing interface
type Storage struct {
	Coupons []Coupon `json:"coupons"`
	Stores  []Store  `json:"stores"`
}

//NewStorage returns an acess to Storage facilites
func NewStorage() Storage {
	return Storage{
		Coupons: make([]Coupon, 0, 20),
		Stores:  make([]Store, 0, 20),
	}
}

//SingleCoupon returns the coupon with the given id
func (s *Storage) SingleCoupon(couponID string) (*listing.SingleCouponResponse, error) {
	v := s.Coupons[0]
	if v.CouponID == couponID {
		c := listing.Coupon{
			CouponID: v.CouponID,
			Store: listing.Store{
				StoreName:  v.Store.StoreName,
				StoreID:    v.Store.StoreID,
				StoreImage: v.Store.StoreImage,
				ThemeColor: v.ThemeColor,
				Tagline:    "Home of all things books",
			},
			QrCodeURL:    "https://9460aae8fa72.ngrok.io/coupon/qrcode/" + "4",
			IsTextCoupon: v.IsTextCoupon,
			AmountOff:    float64(v.AmountOff),
			Desc:         v.Desc,
			CurrencyCode: v.CurrencyCode,

			ExpiringDate: int(v.ExpiringDate),
			DiscountType: v.DiscountType,
		}
		return &listing.SingleCouponResponse{Coupon: &c}, nil
	}

	return nil, listing.ErrCouponNotFound
}

//PopularCoupons returns a list of popular coupons
func (s *Storage) PopularCoupons() (*listing.CouponListResponse, error) {

	listc := make([]listing.Coupon, 0, 30)
	for _, v := range s.Coupons {
		c := listing.Coupon{
			CouponID: v.CouponID,
			Store: listing.Store{
				StoreName:  v.Store.StoreName,
				StoreID:    v.Store.StoreID,
				StoreImage: v.Store.StoreImage,
				ThemeColor: v.ThemeColor,
			},
			AmountOff:     float64(v.AmountOff),
			PercentageOff: float64(v.AmountOff),
			Desc:          v.Desc,
			IsTextCoupon:  v.IsTextCoupon,
			CurrencyCode:  v.CurrencyCode,
			ExpiringDate:  int(v.ExpiringDate),
			DiscountType:  v.DiscountType,
		}
		listc = append(listc, c)
	}
	return &listing.CouponListResponse{
		LastCouponID:   listc[len(listc)-1].CouponID,
		LastCouponTime: s.Coupons[len(s.Coupons)-1].CreationDate,
		Coupons:        listc,
	}, nil
}

//PopularCouponsWithCat returns a list of popular coupons associated with the category
func (s *Storage) PopularCouponsWithCat(cat string) (*listing.CouponListResponse, error) {

	listc := make([]listing.Coupon, 0, 30)
	for _, v := range s.Coupons {
		c := listing.Coupon{
			CouponID: v.CouponID,
			Store: listing.Store{
				StoreName:  v.Store.StoreName,
				StoreID:    v.Store.StoreID,
				StoreImage: v.Store.StoreImage,
				ThemeColor: v.ThemeColor,
			},
			AmountOff:      float64(v.AmountOff),
			Desc:           v.Desc,
			CurrencyCode:   v.CurrencyCode,
			ExpiringDate:   int(v.ExpiringDate),
			IsTextCoupon:   v.IsTextCoupon,
			TextCouponCode: v.TextCouponCode,
			QrCodeURL:      v.QrCodeURL,

			DiscountType: v.DiscountType,
		}
		listc = append(listc, c)
	}
	return &listing.CouponListResponse{
		LastCouponID:   listc[len(listc)-1].CouponID,
		LastCouponTime: s.Coupons[len(s.Coupons)-1].CreationDate,
		Coupons:        listc,
	}, nil
}

//LatestCoupons returns the list of latest coupons
func (s *Storage) LatestCoupons() (*listing.CouponListResponse, error) {
	fmt.Println("latest coupons")
	listc := make([]listing.Coupon, 0, 30)
	for _, v := range s.Coupons {
		c := listing.Coupon{
			CouponID: v.CouponID,
			Store: listing.Store{
				StoreName:  v.Store.StoreName,
				StoreID:    v.Store.StoreID,
				StoreImage: v.Store.StoreImage,
				ThemeColor: v.ThemeColor,
			},
			AmountOff:      float64(v.AmountOff),
			Desc:           v.Desc,
			CurrencyCode:   v.CurrencyCode,
			ExpiringDate:   int(v.ExpiringDate),
			IsTextCoupon:   v.IsTextCoupon,
			TextCouponCode: v.TextCouponCode,
			QrCodeURL:      v.QrCodeURL,
			DiscountType:   v.DiscountType,
		}
		listc = append(listc, c)
	}

	return &listing.CouponListResponse{
		LastCouponID:   listc[len(listc)-1].CouponID,
		LastCouponTime: s.Coupons[len(s.Coupons)-1].CreationDate,
		Coupons:        listc,
	}, nil
}

//TopCategoriesList returns a list of popular Categories associated with coupons
func (s *Storage) TopCategoriesList(limit int) ([]string, error) {
	return []string{"restaurant", "bookstore", "manga", "food", "cinemas"}, nil
}

//CategoriesList returns a list of Categories associated with coupons
func (s *Storage) CategoriesList(limit int) ([]string, error) {
	return []string{"restaurant", "bookstore", "manga", "food", "cinemas"}, nil
}

//LatestCouponWithFiltering returns a list of coupons that has been filtered by a set of filters
func (s *Storage) LatestCouponWithFiltering(filters ...string) (*listing.CouponListResponse, error) {
	fmt.Println("latest coupons with filtering ...")
	listc := make([]listing.Coupon, 0, 30)
	for _, v := range s.Coupons {
		c := listing.Coupon{
			CouponID: v.CouponID,
			Store: listing.Store{
				StoreName:  v.Store.StoreName,
				StoreID:    v.Store.StoreID,
				StoreImage: v.Store.StoreImage,
				ThemeColor: v.ThemeColor,
			},
			IsTextCoupon: false,

			AmountOff:    float64(v.AmountOff),
			Desc:         v.Desc,
			CurrencyCode: v.CurrencyCode,
			ExpiringDate: int(v.ExpiringDate),
			DiscountType: v.DiscountType,
		}
		listc = append(listc, c)
	}
	return &listing.CouponListResponse{
		LastCouponID:   listc[len(listc)-1].CouponID,
		LastCouponTime: s.Coupons[len(s.Coupons)-1].CreationDate,
		Coupons:        listc,
	}, nil
}

//CouponBeforeIDAndTimeFiltered returns the coupons before the last coupon sent to the server with filtered applied is supplied
func (s *Storage) CouponBeforeIDAndTimeFiltered(couponID string, lastTime string, filters ...string) (*listing.CouponListResponse, error) {
	fmt.Println("coupons before id and filtering .....")
	listc := make([]listing.Coupon, 0, 30)
	for _, v := range s.Coupons {
		c := listing.Coupon{
			CouponID: v.CouponID,
			Store: listing.Store{
				StoreName:  v.Store.StoreName,
				StoreID:    v.Store.StoreID,
				StoreImage: v.Store.StoreImage,
				ThemeColor: v.ThemeColor,
			},
			AmountOff:      float64(v.AmountOff),
			Desc:           v.Desc,
			CurrencyCode:   v.CurrencyCode,
			IsTextCoupon:   v.IsTextCoupon,
			TextCouponCode: v.TextCouponCode,
			QrCodeURL:      v.QrCodeURL,
			ExpiringDate:   int(v.ExpiringDate),
			DiscountType:   v.DiscountType,
		}
		listc = append(listc, c)
	}
	return &listing.CouponListResponse{
		LastCouponID:   listc[len(listc)-1].CouponID,
		LastCouponTime: s.Coupons[len(s.Coupons)-1].CreationDate,
		Coupons:        listc,
	}, nil

}

//CouponsSavedByUser returns a list of coupons that has been saved by a user
func (s *Storage) CouponsSavedByUser(userid string) (*listing.CouponListResponse, error) {

	listc := make([]listing.Coupon, 0, 30)
	for _, v := range s.Coupons {
		c := listing.Coupon{
			CouponID: v.CouponID,
			Store: listing.Store{
				StoreName:  v.Store.StoreName,
				StoreID:    v.Store.StoreID,
				StoreImage: v.Store.StoreImage,
				ThemeColor: v.ThemeColor,
			},
			AmountOff:      float64(v.AmountOff),
			Desc:           v.Desc,
			CurrencyCode:   v.CurrencyCode,
			IsTextCoupon:   v.IsTextCoupon,
			TextCouponCode: v.TextCouponCode,
			ExpiringDate:   int(v.ExpiringDate),
			DiscountType:   v.DiscountType,
		}
		listc = append(listc, c)
	}
	return &listing.CouponListResponse{
		LastCouponID:   listc[len(listc)-1].CouponID,
		LastCouponTime: s.Coupons[len(s.Coupons)-1].CreationDate,
		Coupons:        listc,
	}, nil

}

//StoresFollowedByUser returns a list of stores followed by a user
func (s *Storage) StoresFollowedByUser(userid string) (*listing.StoreListResponse, error) {
	st := s.Stores[0]
	store := listing.Store{
		StoreID:     st.StoreID,
		StoreName:   "Rive Global Store",
		StoreImage:  st.StoreImage,
		Address:     st.Address,
		Following:   st.Following,
		Tagline:     st.Tagline,
		CouponCount: st.CouponCount,
	}
	return &listing.StoreListResponse{Stores: []listing.Store{store, store, store}}, nil
}

//UserStoreData contains the data of the user store
func (s *Storage) UserStoreData(userid string) (*storemanagement.UserStoreResponse, error) {
	st := s.Stores[0]
	store := storemanagement.Store{
		StoreID:     st.StoreID,
		StoreName:   "Rive Global Store",
		StoreImage:  st.StoreImage,
		Address:     st.Address,
		Tagline:     st.Tagline,
		CouponCount: st.CouponCount,
	}
	return &storemanagement.UserStoreResponse{
		Store: store,
	}, nil
}

//StoreDetails returns the details of the store associated with a user idI
func (s *Storage) StoreDetails(storeid string) (*listing.SingleStoreResponse, error) {
	if storeid != "123asdf" {
		return nil, listing.ErrStoreNotExists
	}
	st := s.Stores[0]
	store := listing.Store{
		StoreID:     st.StoreID,
		StoreName:   "Rive Global Store",
		StoreImage:  st.StoreImage,
		Address:     st.Address,
		Following:   st.Following,
		Tagline:     st.Tagline,
		CouponCount: st.CouponCount,
	}
	return &listing.SingleStoreResponse{Store: store}, nil
}

func createListingCouponsResponse(c []Coupon) *listing.CouponListResponse {
	listc := make([]listing.Coupon, 0, 20)
	for _, v := range c {
		c := listing.Coupon{
			CouponID: v.CouponID,
			Store: listing.Store{
				StoreName:  "Rive Global Store",
				StoreID:    v.Store.StoreID,
				StoreImage: v.Store.StoreImage,
				ThemeColor: v.ThemeColor,
			},
			AmountOff: float64(v.AmountOff),
			Desc:      v.Desc,
			//Currency:       v.CurrencyCode,
			ExpiringDate:   int(v.ExpiringDate),
			IsTextCoupon:   v.IsTextCoupon,
			TextCouponCode: v.TextCouponCode,
			//	QrCodeURL:        v.QrCodeURL,
			DiscountType: v.DiscountType,
		}
		listc = append(listc, c)
	}
	return &listing.CouponListResponse{Coupons: listc, LastCouponID: c[len(listc)-1].CouponID,
		LastCouponTime: c[len(listc)-1].CreationDate}

}
func createSmanagementCouponsResponse(c []Coupon) *storemanagement.CouponListResponse {
	listc := make([]storemanagement.Coupon, 0, 20)
	for _, v := range c {
		c := storemanagement.Coupon{
			CouponID: v.CouponID,
			Store: storemanagement.Store{
				StoreID:   v.Store.StoreName,
				StoreName: "Rive Global Store",
			},
			AmountOff:      float64(v.AmountOff),
			Desc:           v.Desc,
			CurrencyCode:   v.CurrencyCode,
			ExpiringDate:   int(v.ExpiringDate),
			IsTextCoupon:   v.IsTextCoupon,
			TextCouponCode: v.TextCouponCode,
			QrCodeURL:      v.QrCodeURL,
			DiscountType:   v.DiscountType,
		}
		listc = append(listc, c)
	}
	return &storemanagement.CouponListResponse{Coupons: listc, LastCouponID: c[len(listc)-1].CouponID,
		LastCouponTime: c[len(listc)-1].CreationDate}

}

//StoreCoupons returns the coupons of the store associated with the id
func (s *Storage) StoreCoupons(storeid string) (*listing.CouponListResponse, error) {
	if storeid != "123asdf" {
		return nil, listing.ErrStoreNotExists
	}

	res := createListingCouponsResponse(s.Coupons)
	res.LastCouponID = s.Coupons[len(s.Coupons)-1].CouponID
	res.LastCouponTime = s.Coupons[len(s.Coupons)-1].CreationDate
	return res, nil
}

//CategoryCoupons returns the coupons associated with this category
func (s *Storage) CategoryCoupons(catName string) (*listing.CouponListResponse, error) {

	res := createListingCouponsResponse(s.Coupons)
	res.LastCouponID = s.Coupons[len(s.Coupons)-1].CouponID
	res.LastCouponTime = s.Coupons[len(s.Coupons)-1].CreationDate
	return res, nil
}

//CategoryCouponsBeforeIDAndTime returns the coupons associated with this category
func (s *Storage) CategoryCouponsBeforeIDAndTime(catName, couponid, lastTime string) (*listing.CouponListResponse, error) {

	res := createListingCouponsResponse(s.Coupons)
	res.LastCouponID = s.Coupons[len(s.Coupons)-1].CouponID
	res.LastCouponTime = s.Coupons[len(s.Coupons)-1].CreationDate
	return res, nil
}

//UserStoreCoupons returns the coupons of the store associated with the id
func (s *Storage) UserStoreCoupons(userid string) (*storemanagement.CouponListResponse, error) {

	res := createSmanagementCouponsResponse(s.Coupons)
	res.LastCouponID = s.Coupons[len(s.Coupons)-1].CouponID
	res.LastCouponTime = s.Coupons[len(s.Coupons)-1].CreationDate
	return res, nil
}

//GetUserStoreCouponsRedeemedCount returns the total redeemed count with the filter applied
func (s *Storage) GetUserStoreCouponsRedeemedCount(storeID, filter string) (uint, error) {
	return 3000, nil
}

//CreateCoupon create a coupon for the store
func (s *Storage) CreateCoupon(storeID string, coupon storemanagement.CreateCoupon) (string, error) {
	return s.Coupons[0].CouponID, nil
}

//Employees returns a list of employees associated with the store
func (s *Storage) Employees(storeID string) (*storemanagement.EmployeesResponse, error) {
	Employees := []storemanagement.Employee{
		{FullName: "Adewoyin David", ID: "asdfgh", State: "active", Email: "david Adewoyin hotmail.com"},
		{FullName: "Betty Trotwood", ID: "asdfgh", State: "suspended", Email: "david Adewoyin hotmail.com"},
		{FullName: "David CopperField", ID: "asdfgh", State: "active", Email: "david Adewoyin hotmail.com"},
		{FullName: "Robinsoe Crusoe", ID: "asdfgh", State: "suspended", Email: "david Adewoyin hotmail.com"},
		{FullName: "Uzumaki Naruto", ID: "asdfgh", State: "suspended", Email: "david Adewoyin hotmail.com"},
		{FullName: "Juniper lee", ID: "asdfgh", State: "active", Email: "david Adewoyin hotmail.com"},
		{FullName: "Martin Morning", ID: "asdfgh", State: "suspended", Email: "david Adewoyin hotmail.com"},
		{FullName: "Alice greenfinger", ID: "asdfgh", State: "active", Email: "david Adewoyin hotmail.com"},
		{FullName: "Jet li", ID: "asdfgh", State: "suspended", Email: "david Adewoyin hotmail.com"},
		{FullName: "Bryan haggins", ID: "asdfgh", State: "active", Email: "david Adewoyin hotmail.com"},
	}
	fmt.Println("Here now david")
	return &storemanagement.EmployeesResponse{
		Employees: Employees,
	}, nil
}

//AddEmployee register an employee under the given store
func (s *Storage) AddEmployee(storeID string, employee storemanagement.Employee) error { return nil }

//SuspendEmployee suspends the given  store employee
func (s *Storage) SuspendEmployee(storeID string, employeeID string) error { return nil }

//RemoveEmployee removes the given employee from the store
func (s *Storage) RemoveEmployee(storeID string, employeeID string) error { return nil }

//ResumeEmployee resumes the suspended employee
func (s *Storage) ResumeEmployee(storeID string, employeeID string) error { return nil }

//SubStores returns a list of sub stores associated with the store
func (s *Storage) SubStores(storeID string) ([]storemanagement.Store, error) {
	return []storemanagement.Store{
		{StoreID: "123asf", StoreName: "Rive Nigeria"},
		{StoreID: "123asf", StoreName: "Rive Global"},
		{StoreID: "123asf", StoreName: "Rive Michigan"},
	}, nil
}

//EditStore edit the details if a particular store
func (s *Storage) EditStore(userID string, store storemanagement.StoreEdit) error { return nil }

//AddSubStore adds a store as a substore for a given store
func (s *Storage) AddSubStore(storeID string, otherStoreID string) error { return nil }

//RemoveSubStore removes a store for the list of the substores associated with a store
func (s *Storage) RemoveSubStore(storeid string, subStoreID string) error { return nil }

//CreateUser create a user with the given details and returns the userid
func (s *Storage) CreateUser(email string, password string) (bool, error) {
	return true, nil
}

//UserWithEmail returns the userid asscoiated with the user or an error
func (s *Storage) UserWithEmail(email string) (string, error) {
	if email == "davidadewoyin@hotmail.com" {
		return "123", nil
	}
	return "", errors.New("user already exists")
}

//TODO add password hashing
//TODO hash fake password to stop delay for wrong email address
//UserWithEmailAndPassword returns the userid associated with the email and password
func (s *Storage) UserWithEmailAndPassword(email, password string) (string, error) {

	if email == "davidadewoyin@hotmail.com" && password == "password" {
		return "123", nil
	}
	return "", errors.New("user not found")
}

//TODO
//ResetPassword resets the password of the indentity if it exists
func (s *Storage) ResetPassword(email string) error {
	return nil
}

//DeleteCoupon deletes the coupon from the store
func (s *Storage) DeleteCoupon(storeID string, couponID string) error {
	return nil
}

//CouponsSavedByUserBeforeIDAndTime returns a list of coupons before the specified id and time
func (s *Storage) CouponsSavedByUserBeforeIDAndTime(userid, couponid, lastTime string) (*listing.CouponListResponse, error) {
	listc := make([]listing.Coupon, 0, 30)
	for _, v := range s.Coupons {
		c := listing.Coupon{
			CouponID: v.CouponID,
			Store: listing.Store{
				StoreName:  v.Store.StoreName,
				StoreID:    v.Store.StoreID,
				StoreImage: v.Store.StoreImage,
				ThemeColor: v.ThemeColor,
			},
			AmountOff:      float64(v.AmountOff),
			Desc:           v.Desc,
			CurrencyCode:   v.CurrencyCode,
			IsTextCoupon:   v.IsTextCoupon,
			TextCouponCode: v.TextCouponCode,
			QrCodeURL:      v.QrCodeURL,
			ExpiringDate:   int(v.ExpiringDate),
			DiscountType:   v.DiscountType,
		}
		listc = append(listc, c)
	}
	return &listing.CouponListResponse{
		LastCouponID:   listc[len(listc)-1].CouponID,
		LastCouponTime: s.Coupons[len(s.Coupons)-1].CreationDate,
		Coupons:        listc,
	}, nil
}

//StoreCouponsBeforeIDAndTime returns a list of coupons before the specified id and time
func (s *Storage) StoreCouponsBeforeIDAndTime(userid, couponid, lastTime string) (*listing.CouponListResponse, error) {
	return nil, nil
}

//QrImage returns the qrimage of the coupon
func (s *Storage) QrImage(imageID string) ([]byte, error) {
	return s.GenerateCouponImage("Hello David"), nil
}

//CouponState returns the state of the coupon
func (s *Storage) CouponState(couponid string) (string, error) {
	return storemanagement.CouponActive, nil
}

//VerifyCoupon verifies the coupon
func (s *Storage) VerifyCoupon(userid, couponid string) error {
	return nil
}

//GenerateCouponImage generate the qr code of the coupon
func (s *Storage) GenerateCouponImage(desc string) []byte {
	couponid := uuid.New().String()
	time := time.Now().Unix()

	type Payload struct {
		Id   string `json:"id,omitempty"`
		Exp  int64  `json:"exp,omitempty"`
		Text string `json:"text,omitempty"`
	}

	type QrImagePayload struct {
		Sig  string  `json:"sig,omitempty"`
		Data Payload `json:"data,omitempty"`
	}

	coupon := QrImagePayload{Sig: "HleWj5EGbZhantO8hMknienvaLybNldSlB64ZZ7W1qIOCKKP9omm93Tl6F64sdYFNCepEBQUga9nEW90xtT3fY7vJXFFHytG1J3n0FTcqvRwQpXxwPzxLgyf8MMXccQai4OaXUH5KGPjx3n8N5PWNZ19uZCbBAg59a1Ei0RM",
		Data: Payload{Id: couponid, Exp: time, Text: "VCAuxcbxfUIZKg2eQts6HDRIBcHyW23pvvDnVZ5KqHxDeOOjgcK81cLLltUZsoVBUzspI9TLgGPKGIC7HVhCvtC2534rm5M1kQ50NM0ppERxL3cHJtWquV3POPsiv"}}
	/* coupon := QrImagePayload{Sig: "HleWj5EGbZhantO8hMknienvaLybNldSlB64ZZ7W1qIOZ19uZCbBAg59a1Ei0RM",
	Data: Payload{Id: couponid, Exp: time, Text: "VCAuxcbxfUIZKg2eQts6HDRIBcHyW23pvvDnVZ5KqHxDesiv"}} */
	cJSONByte, err := json.Marshal(coupon)
	if err != nil {
		fmt.Println(err)
	}
	cJSON := string(cJSONByte)
	fmt.Println(cJSON)
	fmt.Println(len(cJSON))

	value, err := qrcode.New(cJSON, qrcode.Medium)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	/* value.ForegroundColor = color.NRGBA{
		R: 255,
		G: 108,
		B: 135,
		A: 255,
	}
	value.BackgroundColor = color.NRGBA{
		R: 18,
		G: 18,
		B: 18,
		A: 28,
	} */
	v := value.Image(300)
	f, _ := os.Create("dddaa.png")
	png.Encode(f, v)

	qrbytes, err := value.PNG(300)
	if err != nil {
		fmt.Println(err)
	}
	return qrbytes

}
