package listing

import (
	"context"
	"errors"
)

var (
	//ErrUserNotExists is returned if a user does not exists with the given userid
	ErrUserNotExists = errors.New("user does not exists")
	//ErrStoreNotFound is returned if a store does not exists with the given storeid
	ErrStoreNotFound = errors.New("store not found")
	//ErrCouponsNotFound  is returned if no coupons is associated with the given id
	ErrCouponsNotFound = errors.New("coupons not found")
	//ErrCouponNotFound  is returned if no coupon is associated with the given id
	ErrCouponNotFound = errors.New("coupons not found")
	//ErrStoresNotFound is returned if no stores can be returned
	ErrStoresNotFound = errors.New("stores not found")
)

//MaxRedemption represent the limit the maximum  number a coupon should be redeem
type MaxRedemption string

//CouponListResponse is sent to the client containing the coupons data
type CouponListResponse struct {
	LastCouponID string   `json:"last_coupon_id,omitempty"`
	Coupons      []Coupon `json:"coupons,omitempty"`
}

//SingleCouponResponse is sent to the client containing the data of a single coupon
type SingleCouponResponse struct {
	Coupon *Coupon `json:"coupon,omitempty"`
}

//SingleStoreResponse is sent to the client containing data of a single store
type SingleStoreResponse struct {
	Store Store `json:"store,omitempty"`
}

//StoreListResponse is sent to the client containing the stores data
type StoreListResponse struct {
	Stores []Store `json:"stores,omitempty"`
}

//CategoriesResponse is returned containing the list of filter Categories
type CategoriesResponse struct {
	Categories []string `json:"categories,omitempty"`
}

// Coupon represent a redeemable piece of text or qrcode
type Coupon struct {
	CouponID         string   `json:"coupon_id,omitempty"`
	Store            Store    `json:"store,omitempty"`
	AmountOff        float64  `json:"amount_off,omitempty"`
	PercentageOff    float64  `json:"percentage_off,omitempty"`
	Desc             string   `json:"desc,omitempty"`
	CurrencyCode     string   `json:"currency_code,omitempty"`
	ExpiringDate     int      `json:"expiring_date,omitempty"`
	Categories       []string `json:"Categories,omitempty"`
	IsTextCoupon     bool     `json:"is_text_coupon,omitempty"`
	TextCouponCode   string   `json:"text_coupon_code,omitempty"`
	WebUrl           string   `json:"web_url,omitempty"`
	TextCouponWebURL string   `json:"text_coupon_web_url,omitempty"`
	QrCodeURL        string   `json:"qr_code_url,omitempty"`
	DiscountType     string   `json:"discount_type,omitempty"`
	// voucher percentage
}

// Store represent an indentity that owns coupons
type Store struct {
	StoreID     string `json:"store_id,omitempty"`
	StoreName   string `json:"store_name,omitempty"`
	ThemeColor  int    `json:"theme_color,omitempty"`
	StoreImage  string `json:"store_image,omitempty"`
	Tagline     string `json:"tagline,omitempty"`
	Address     string `json:"address,omitempty"`
	Following   bool   `json:"following,omitempty"`
	CouponCount uint   `json:"coupon_count,omitempty"`
}

// User is an identifiable entity
type User struct {
	UserID   string
	Fullname string
}

//Repository provides access to coupon and store storage facilities
type Repository interface {
	SingleCoupon(ctx context.Context, couponID string) (*SingleCouponResponse, error)
	PopularCoupons(ctx context.Context) (*CouponListResponse, error)
	PopularCouponsWithFiltering(ctx context.Context, cat string) (*CouponListResponse, error)
	LatestCoupons(ctx context.Context) (*CouponListResponse, error)
	LatestCouponsWithFiltering(ctx context.Context, filter string) (*CouponListResponse, error)

	TopCategories(ctx context.Context, limit int) ([]string, error)
	Categories(ctx context.Context) ([]string, error)
	CategoryCoupons(ctx context.Context, catName string) (*CouponListResponse, error)
	CategoryCouponsBeforeIDAndTime(ctx context.Context, catName, couponid, lastTime string) (*CouponListResponse, error)

	CouponBeforeIDAndTimeFiltered(ctx context.Context, couponid string, lastTime string, filter string) (*CouponListResponse, error)
	CouponsSavedByUser(ctx context.Context, userid string) (*CouponListResponse, error)
	/* CouponsSavedByUserBeforeIDAndTime(ctx context.Context, userid, couponid, lastTime string) (*CouponListResponse, error) */
	StoresFollowedByUser(ctx context.Context, userid string) (*StoreListResponse, error)
	StoreDetails(ctx context.Context, storeid string) (*SingleStoreResponse, error)
	StoreCoupons(ctx context.Context, storeid string) (*CouponListResponse, error)
	StoreCouponsBeforeIDAndTime(ctx context.Context, storeid, couponid, lastTime string) (*CouponListResponse, error)
	QrImage(ctx context.Context, image string) ([]byte, error)
}

// Service provides store and coupon listing operation
type Service interface {
	SingleCoupon(ctx context.Context, couponID string) (*SingleCouponResponse, error)
	PopularCoupons(ctx context.Context) (*CouponListResponse, error)
	PopularCouponsWithFiltering(ctx context.Context, cat string) (*CouponListResponse, error)
	LatestCoupons(ctx context.Context) (*CouponListResponse, error)
	LatestCouponsWithFiltering(ctx context.Context, filter string) (*CouponListResponse, error)

	TopCategories(ctx context.Context, limit int) (*CategoriesResponse, error)
	Categories(ctx context.Context) (*CategoriesResponse, error)

	CouponBeforeIDAndTimeFiltered(ctx context.Context, couponid string, lastTime string, filter string) (*CouponListResponse, error)

	CouponsSavedByUser(ctx context.Context, userid string) (*CouponListResponse, error)
	/* CouponsSavedByUserBeforeIDAndTime(ctx context.Context, userid, couponid, lastTime string) (*CouponListResponse, error) */
	StoresFollowedByUser(ctx context.Context, userid string) (*StoreListResponse, error)
	StoreDetails(ctx context.Context, storeid string) (*SingleStoreResponse, error)
	StoreCoupons(ctx context.Context, storeid string) (*CouponListResponse, error)
	StoreCouponsBeforeIDAndTime(ctx context.Context, storeid, couponid, lastTime string) (*CouponListResponse, error)
	CategoryCoupons(ctx context.Context, catName string) (*CouponListResponse, error)
	CategoryCouponsBeforeIDAndTime(ctx context.Context, catName, couponid, lastTime string) (*CouponListResponse, error)
	QrImage(ctx context.Context, image string) ([]byte, error)
}
type service struct {
	r Repository
}

// NewService creates a listing service
func NewService(repo Repository) Service {
	return &service{r: repo}
}

func (s *service) TopCategories(ctx context.Context, limit int) (*CategoriesResponse, error) {
	Categories, err := s.r.TopCategories(ctx, limit)
	if err != nil {
		return nil, errors.New("internal server errror")
	}

	return &CategoriesResponse{Categories: Categories}, nil
}
func (s *service) Categories(ctx context.Context) (*CategoriesResponse, error) {
	Categories, err := s.r.Categories(ctx)
	if err != nil {
		return nil, errors.New("internal server errror")
	}

	return &CategoriesResponse{Categories: Categories}, nil
}

func (s *service) SingleCoupon(ctx context.Context, couponid string) (*SingleCouponResponse, error) {
	return s.r.SingleCoupon(ctx, couponid)

}
func (s *service) PopularCoupons(ctx context.Context) (*CouponListResponse, error) {
	return s.r.PopularCoupons(ctx)
}
func (s *service) PopularCouponsWithFiltering(ctx context.Context, cat string) (*CouponListResponse, error) {
	return s.r.PopularCouponsWithFiltering(ctx, cat)
}

func (s *service) LatestCoupons(ctx context.Context) (*CouponListResponse, error) {
	return s.r.LatestCoupons(ctx)

}
func (s *service) LatestCouponsWithFiltering(ctx context.Context, filter string) (*CouponListResponse, error) {
	return s.r.LatestCouponsWithFiltering(ctx, filter)
}
func (s *service) CouponsSavedByUser(ctx context.Context, userid string) (*CouponListResponse, error) {
	return s.r.CouponsSavedByUser(ctx, userid)
}

/* func (s *service) CouponsSavedByUserBeforeIDAndTime(ctx context.Context, userid string, couponid, lastTime string) (*CouponListResponse, error) {
	return s.r.CouponsSavedByUserBeforeIDAndTime(ctx, userid, couponid, lastTime)
} */
func (s *service) StoresFollowedByUser(ctx context.Context, userid string) (*StoreListResponse, error) {
	return s.r.StoresFollowedByUser(ctx, userid)
}
func (s *service) StoreDetails(ctx context.Context, storeid string) (*SingleStoreResponse, error) {
	return s.r.StoreDetails(ctx, storeid)
}
func (s *service) StoreCoupons(ctx context.Context, storeid string) (*CouponListResponse, error) {
	return s.r.StoreCoupons(ctx, storeid)
}
func (s *service) CouponBeforeIDAndTimeFiltered(ctx context.Context, couponid string, lastTime string, filter string) (*CouponListResponse, error) {
	return s.r.CouponBeforeIDAndTimeFiltered(ctx, couponid, lastTime, filter)
}
func (s *service) StoreCouponsBeforeIDAndTime(ctx context.Context, storeid, couponid string, lastTime string) (*CouponListResponse, error) {
	return s.r.StoreCouponsBeforeIDAndTime(ctx, storeid, couponid, lastTime)
}
func (s *service) CategoryCoupons(ctx context.Context, catName string) (*CouponListResponse, error) {
	return s.r.CategoryCoupons(ctx, catName)
}
func (s *service) CategoryCouponsBeforeIDAndTime(ctx context.Context, catName, couponid, lastTime string) (*CouponListResponse, error) {
	return s.r.CategoryCouponsBeforeIDAndTime(ctx, catName, couponid, lastTime)
}
func (s *service) QrImage(ctx context.Context, image string) ([]byte, error) {
	return s.r.QrImage(ctx, image)
}
