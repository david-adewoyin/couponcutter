package storemanagement

import (
	"context"
	"errors"
)

var (
	//ErrUnableToCreateCoupon is returned if an error occurs while creating a coupons
	ErrUnableToCreateCoupon = errors.New("unable to create coupon")
	//ErrUnableToCreateStore is returned if an error occurs while creating a store
	ErrUnableToCreateStore = errors.New("unable to create store")
	//ErrEmployeeOpFails  is returned if an error occurs while performing an operation on an employee
	ErrEmployeeOpFails = errors.New("unable to perform operation on employee")
	//ErrEmployeeNotFound is returned if no employee is associated with the given id
	ErrEmployeeNotFound = errors.New("employee not found")
	//ErrStoreNotFound is returned if no store is associated with the id
	ErrStoreNotFound = errors.New("store not found")
	//ErrCouponNotValid is returned if a couponid is not associated with a coupon
	ErrCouponNotValid = errors.New("coupon is not valid")
	//ErrCouponIsUsed is returned if a coupon is already used
	ErrCouponIsUsed = errors.New("coupon is already used")
	//ErrCouponLimitExceeded is returned if the limit of redemption amount is exceeded
	ErrCouponLimitExceeded = errors.New("coupon limit already exceeded")
)

const (
	//EmployeeActive is the state of an active employee
	EmployeeActive = "active"
	//EmployeeSuspended is the state of a suspended employee
	EmployeeSuspended = "suspended"
	//EmployeeRemoved is the state of an employee to be remove
	EmployeeRemoved = "delete"
	//CouponActive represent a coupon is active for redeemption
	CouponActive = "active"
	//CouponUsed represent a coupon already used or over redemption limit
	CouponUsed = "used"
	//CouponInActive represent a coupon that is no more active
	CouponInActive = "inactive"
)

//UserStoreResponse is returned containing the data of the user store
type UserStoreResponse struct {
	Store Store `json:"store,omitempty"`
}

//EmployeesResponse is returned containing the data of the employees of a store
type EmployeesResponse struct {
	Employees []Employee `json:"employees,omitempty"`
}

//CouponListResponse is returned containing the store coupons of the user
type CouponListResponse struct {
	Coupons        []Coupon `json:"coupons,omitempty"`
	LastCouponID   string   `json:"last_coupon_id,omitempty"`
	LastCouponTime uint     `json:"last_coupon_time,omitempty"`
}

//MaxRedemption represent the limit the maximum  number a coupon should be redeem
type MaxRedemption string

// Coupon represent a redeemable piece of text or qrcode
type Coupon struct {
	CouponID            string  `json:"coupon_id"`
	Store               Store   `json:"store,omitempty"`
	AmountOff           float64 `json:"amount_off,omitempty"`
	PercentageOff       float64 `json:"percentage_off,omitempty"`
	Desc                string  `json:"desc,omitempty"`
	CurrencyCode        string  `json:"currency_code,omitempty"`
	CreationTime        uint    `json:"creation_time,omitempty"`
	State               string  `json:"state,omitempty"`
	ExpiringDate        int     `json:"expiring_date,omitempty"`
	SingleUserUse       bool    `json:"single_user_use,omitempty"`
	DiscountType        string  `json:"discount_type,omitempty"`
	IsTextCoupon        bool    `json:"is_text_coupon,omitempty"`
	TextCouponCode      string  `json:"text_coupon_code,omitempty"`
	QrCodeURL           string  `json:"item_url,omitempty"`
	UnlimitedRedemption bool    `json:"unlimted_redemption,omitempty"`
	MaxRedemption       uint    `json:"redemption_limit,omitempty"`
	// voucher percentage
}

//CreateCoupon is used to create coupons
type CreateCoupon struct {
	StoreID       string   `json:"store_Id,omitempty"`
	AmountOff     float64  `json:"amount_off,omitempty"`
	PercentageOff float64  `json:"percentage_off,omitempty"`
	Desc          string   `json:"desc,omitempty"`
	Categories    []string `json:"categories"`
	CurrencyCode  string   `json:"currency_code,omitempty"`
	ExpiringDate  uint     `json:"expiring_date"`
	State         string   `json:"state,omitempty"`

	//SingleUserUse       bool     `json:"single_user_use,omitempty"`
	IsTextCoupon        bool   `json:"is_text_coupon,omitempty"`
	TextCouponCode      string `json:"text_coupon_code,omitempty"`
	TextCouponWebURL    string `json:"text_coupon_weburl,omitempty"`
	UnlimitedRedemption bool   `json:"unlimted_redemption"`
	MaxRedemption       uint   `json:"redemption_limit,omitempty"`
	DiscountType        string `json:"discount_type,omitempty"`
}

// Store represent an indentity that owns coupons
type Store struct {
	StoreID     string     `json:"store_id,omitempty"`
	StoreName   string     `json:"store_name,omitempty"`
	StoreImage  string     `json:"store_image,omitempty"`
	Tagline     string     `json:"tagline,omitempty"`
	Address     string     `json:"address,omitempty"`
	CouponCount uint       `json:"coupon_count,omitempty"`
	Employees   []Employee `json:"employees,omitempty"`
	SubStores   []Store    `json:"sub_stores,omitempty"`
	Coupons     []Coupon   `json:"coupons,omitempty"`
}

//StoreEdit is used to edit the change in store details
type StoreEdit struct {
	Name    string
	Tagline string
	Address string
}

//Employee is an identifiable entity that has limited store_management abilities
type Employee struct {
	ID       string `json:"id,omitempty"`
	FullName string `json:"full_name,omitempty"`
	Email    string `json:"email,omitempty"`
	State    string `json:"state,omitempty"`
}

// Repository provides acess to store management storage facilities
type Repository interface {
	CreateCoupon(ctx context.Context, storeID string, coupon CreateCoupon) (string, error)
	DeleteCoupon(ctx context.Context, storeID string, couponID string) error

	UserStoreData(ctx context.Context, userid string) (*UserStoreResponse, error)
	UserStoreCoupons(ctx context.Context, userid string) (*CouponListResponse, error)

	Employees(ctx context.Context, storeID string) (*EmployeesResponse, error)
	AddEmployee(ctx context.Context, storeID string, employee Employee) error
	SuspendEmployee(ctx context.Context, storeID string, employeeID string) error
	RemoveEmployee(ctx context.Context, storeID string, employeeID string) error
	ResumeEmployee(ctx context.Context, storeID string, employeeID string) error

	GetUserStoreCouponsRedeemedCount(ctx context.Context, storeID, filter string) (uint, error)
	CouponState(ctx context.Context, couponid string) (string, error)
	VerifyCoupon(ctx context.Context, userID, couponid string) error

	EditStore(ctx context.Context, userid string, edit StoreEdit) error
}

// Service provides store management facilities
type Service interface {
	CreateCoupon(ctx context.Context, storeID string, coupon CreateCoupon) (string, error)
	DeleteCoupon(ctx context.Context, storeID string, couponID string) error

	UserStoreData(ctx context.Context, userid string) (*UserStoreResponse, error)
	Employees(ctx context.Context, storeID string) (*EmployeesResponse, error)
	UserStoreCoupons(ctx context.Context, userid string) (*CouponListResponse, error)

	AddEmployee(ctx context.Context, storeID string, employee Employee) error
	SuspendEmployee(ctx context.Context, storeID string, employeeID string) error
	RemoveEmployee(ctx context.Context, storeID string, employeeID string) error
	ResumeEmployee(ctx context.Context, storeID string, employeeID string) error

	GetUserStoreCouponsRedeemedCount(ctx context.Context, storeID, filter string) (uint, error)

	CouponState(ctx context.Context, couponid string) (string, error)
	VerifyCoupon(ctx context.Context, userID, couponid string) error

	EditStore(ctx context.Context, userid string, edit StoreEdit) error
}

// NewService returns a store management service provider
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

type service struct {
	repo Repository
}

func (s *service) CreateCoupon(ctx context.Context, storeID string, coupon CreateCoupon) (string, error) {
	return s.repo.CreateCoupon(ctx, storeID, coupon)
}
func (s *service) Employees(ctx context.Context, storeID string) (*EmployeesResponse, error) {
	return s.repo.Employees(ctx, storeID)
}

func (s *service) AddEmployee(ctx context.Context, storeID string, employee Employee) error {
	return s.repo.AddEmployee(ctx, storeID, employee)
}
func (s *service) SuspendEmployee(ctx context.Context, storeID string, employeeID string) error {
	return s.repo.SuspendEmployee(ctx, storeID, employeeID)
}
func (s *service) RemoveEmployee(ctx context.Context, storeID string, employeeID string) error {
	return s.repo.RemoveEmployee(ctx, storeID, employeeID)
}
func (s *service) ResumeEmployee(ctx context.Context, storeID string, employeeID string) error {
	return s.repo.RemoveEmployee(ctx, storeID, employeeID)
}

func (s *service) DeleteCoupon(ctx context.Context, storeID string, couponID string) error {
	return s.repo.DeleteCoupon(ctx, storeID, couponID)
}

func (s *service) UserStoreData(ctx context.Context, userid string) (*UserStoreResponse, error) {
	return s.repo.UserStoreData(ctx, userid)
}
func (s *service) UserStoreCoupons(ctx context.Context, userid string) (*CouponListResponse, error) {
	return s.repo.UserStoreCoupons(ctx, userid)
}
func (s *service) GetUserStoreCouponsRedeemedCount(ctx context.Context, storeID, filter string) (uint, error) {
	return s.repo.GetUserStoreCouponsRedeemedCount(ctx, storeID, filter)
}

func (s *service) CouponState(ctx context.Context, couponid string) (string, error) {
	return s.repo.CouponState(ctx, couponid)
}

func (s *service) VerifyCoupon(ctx context.Context, userID, couponid string) error {
	return s.repo.VerifyCoupon(ctx, userID, couponid)
}
func (s *service) EditStore(ctx context.Context, userid string, edit StoreEdit) error {
	return s.repo.EditStore(ctx, userid, edit)
}
