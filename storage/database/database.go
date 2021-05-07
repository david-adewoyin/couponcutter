package database

import (
	"context"
	"couponcutter/authetication"
	"couponcutter/listing"
	"couponcutter/storage"
	"couponcutter/storemanagement"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/hashicorp/go-hclog"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Database struct {
	dbPool *pgxpool.Pool
	logger hclog.Logger
	psql   sq.StatementBuilderType
}

//NewStorage provides an interface for interacting with a postgres database
func NewStorage() (*Database, error) {
	logger := hclog.New(&hclog.LoggerOptions{Name: "database", Level: hclog.Debug})
	connStr := "postgresql://localhost/couponcutter?user=couponcutter&password=david"
	dbPool, err := pgxpool.Connect(context.Background(), connStr)

	if err != nil {
		logger.Error("unable to connect to database", err)
		return nil, err
	}

	st := &Database{
		dbPool: dbPool,
		logger: logger,
		psql:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
	return st, nil
}

//TODO
func (s *Database) ResetPassword(ctx context.Context, email string) error {
	return nil
}
func (s *Database)SearchCoupon(ctx context.Context,term string)(interface{},error){
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		s.logger.Debug(err.Error())
		return "", storage.ErrServerError
	}
c:=`coupon_id,store_id,store_id,store_name,tagline,theme_color,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl from coupons where tokens @@ $1`
	rows,err:= conn.Query(ctx,c,term)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response listing.CouponListResponse
	for rows.Next() {
		var coupon listing.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Store.ThemeColor,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.TextCouponWebURL,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
		response.LastCouponID = response.Coupons[len(response.Coupons)-1].CouponID
}
func (s *Database) UserWithEmail(ctx context.Context, email string) (string, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		s.logger.Debug(err.Error())
		return "", storage.ErrServerError
	}
	c := s.psql.Select(`user_id`).From("users").Where(sq.Eq{"email": ""})

	cStr, _, err := c.ToSql()
	if err != nil {
		return "", storage.ErrServerError
	}
	row := conn.QueryRow(ctx, cStr, email)

	var userid string
	err = row.Scan(
		&userid,
	)
	if err != nil {
		s.logger.Error(err.Error())
		if errors.Is(err, pgx.ErrNoRows) {
			return "", listing.ErrUserNotExists
		}
		return "", storage.ErrServerError
	}

	return userid, nil

}

//create a user with the given password
func (s *Database) CreateUser(ctx context.Context, email, password string) (string, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		s.logger.Debug(err.Error())
		return "", storage.ErrServerError
	}
	hash, err := storage.HashPassword(password)
	if err != nil {
		s.logger.Error(err.Error())
		return "", storage.ErrServerError
	}

	token := storage.GenerateUUID()
	expired_at := time.Now().Add(time.Hour * 24 * 7)

	r := conn.QueryRow(ctx, "select * from users where email = $1", email)
	// a returned value means a user already exists with the given email
	e := r.Scan()
	if e != nil {
		return "", authetication.ErrIdentityAlreadyExists
	}

	_, err = conn.Exec(ctx, "insert into signup_users(email,password_hash,token,expired_at)values($1 ,$2)", email, password, expired_at, hash, token)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Message) // => syntax error at end of input
			fmt.Println(pgErr.Code)    // => 42601
		}
		s.logger.Error(err.Error())
		return "", storage.ErrServerError
	}

	return token, nil

}

//Fetch user id with the given email and password
func (s *Database) UserWithIdentity(ctx context.Context, email, password string) (string, error) {

	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		s.logger.Debug(err.Error())
		return "", storage.ErrServerError
	}
	hash, err := storage.HashPassword(password)
	if err != nil {
		s.logger.Debug(err.Error())
		return "", storage.ErrServerError
	}

	var userid string
	var passwdHash string

	r := conn.QueryRow(ctx, "select user_id,password_hash from users where email = $1", email)
	err = r.Scan(&userid, &passwdHash)
	if err != nil {
		s.logger.Debug(err.Error())
		return "", authetication.ErrIdentityDoesNotExists
	}
	if passwdHash != hash {
		return "", authetication.ErrIdentityAlreadyExists
	}

	return userid, nil

}

//fetch the coupon details associated with the given id
func (s *Database) SingleCoupon(ctx context.Context, couponId string) (*listing.SingleCouponResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}

	c := s.psql.Select(`coupon_id,store_id,store_id,store_name,tagline,theme_color,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl`).From("coupons").InnerJoin("stores using(store_id)").Where(sq.Eq{"coupon_id": "", "coupon_state": ""})

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	row := conn.QueryRow(ctx, cStr, couponId, "active")

	var couponResponse listing.SingleCouponResponse
	couponResponse.Coupon = &listing.Coupon{}
	err = row.Scan(
		&couponResponse.Coupon.CouponID,
		&couponResponse.Coupon.Store.StoreID,
		&couponResponse.Coupon.Store.StoreName,
		&couponResponse.Coupon.Store.Tagline,
		&couponResponse.Coupon.Store.Address,
		&couponResponse.Coupon.Store.ThemeColor,
		&couponResponse.Coupon.Desc,
		&couponResponse.Coupon.AmountOff,
		&couponResponse.Coupon.PercentageOff,
		&couponResponse.Coupon.CurrencyCode,
		&couponResponse.Coupon.QrCodeURL,
		&couponResponse.Coupon.ExpiringDate,
		&couponResponse.Coupon.IsTextCoupon,
		&couponResponse.Coupon.TextCouponCode,
		&couponResponse.Coupon.TextCouponWebURL,
	)
	if err != nil {
		s.logger.Error(err.Error())
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, listing.ErrCouponNotFound
		}
		return nil, storage.ErrServerError
	}

	return &couponResponse, nil
}

//fetch a list of the most popular coupons
func (s *Database) PopularCoupons(ctx context.Context) (*listing.CouponListResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}

	c := s.psql.Select(`coupon_id,store_id,store_id,store_name,tagline,theme_color,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl`).From("popular_coupons").OrderBy("coupon_id desc")

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, cStr)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response listing.CouponListResponse
	for rows.Next() {
		var coupon listing.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Store.ThemeColor,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.TextCouponWebURL,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
		response.LastCouponID = response.Coupons[len(response.Coupons)-1].CouponID

	}
	if len(response.Coupons) == 0 {
		return nil, listing.ErrCouponsNotFound
	}

	return &response, nil
}

//fetch a list of the  latest coupons
func (s *Database) LatestCoupons(ctx context.Context) (*listing.CouponListResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}

	c := s.psql.Select(`coupon_id,store_id,store_id,store_name,tagline,theme_color,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl`).From("coupons").InnerJoin("stores using (store_id)").Where("coupon_state='active'").OrderBy("created_at desc")

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, cStr)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response listing.CouponListResponse
	for rows.Next() {
		var coupon listing.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Store.ThemeColor,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.TextCouponWebURL,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
		response.LastCouponID = response.Coupons[len(response.Coupons)-1].CouponID

	}
	if len(response.Coupons) == 0 {
		return nil, listing.ErrCouponsNotFound
	}

	return &response, nil

}

//fetch a list of  popular coupons in a particular category
func (s *Database) PopularCouponsWithFiltering(ctx context.Context, catName string) (*listing.CouponListResponse, error) {

	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}
	subQ := s.psql.Select("*").From("popular_coupons").InnerJoin("categories using (cat_id)").Where(sq.Eq{"cat_name": ""})

	c := s.psql.Select(`coupon_id,store_id,store_id,store_name,tagline,theme_color,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl`).FromSelect(subQ, "coupons")

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, cStr, catName)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response listing.CouponListResponse
	for rows.Next() {
		var coupon listing.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Store.ThemeColor,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.TextCouponWebURL,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
		response.LastCouponID = response.Coupons[len(response.Coupons)-1].CouponID
	}
	if len(response.Coupons) == 0 {
		return nil, listing.ErrCouponsNotFound
	}
	return &response, nil
}

func (s *Database) LatestCouponsWithFiltering(ctx context.Context, filter string) (*listing.CouponListResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}
	cQ := s.psql.Select("select cat_id from categories").InnerJoin("coupon_categories using(cat_id)").Where(sq.Eq{"cat_name": ""})
	cQ = s.psql.Select("*").FromSelect(cQ, "cats").InnerJoin("coupons using(coupon_id)")
	stQ := s.psql.Select("*").FromSelect(cQ, "stores").InnerJoin("stores using(store_id)").OrderBy("coupon_id desc")

	qStr, _, err := stQ.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}

	rows, err := conn.Query(ctx, qStr, filter)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}

	var response listing.CouponListResponse
	for rows.Next() {
		var coupon listing.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Store.ThemeColor,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.TextCouponWebURL,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
		response.LastCouponID = response.Coupons[len(response.Coupons)-1].CouponID
	}
	if len(response.Coupons) == 0 {
		return nil, listing.ErrCouponsNotFound
	}
	return &response, nil

}

//fetch a list of the top  coupon categories
func (s *Database) TopCategories(ctx context.Context, limit int) ([]string, error) {
	if limit < 10 {
		limit = 10
	}

	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}
	cQ := s.psql.Select("cat_name,count(*)").From("categories").InnerJoin("coupon_categories using(cat_id)").GroupBy("cat_id").OrderBy("count(*) desc").Limit(uint64(limit))

	cStr, _, err := cQ.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, cStr)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response = make([]string, 0, 20)
	var cat string
	for rows.Next() {
		err = rows.Scan(
			&cat,
		)
		if err != nil {
			continue
		}
		response = append(response, cat)
	}
	return response, nil
}

//fetch a list of the coupon categories
func (s *Database) Categories(ctx context.Context) ([]string, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}
	cQ := s.psql.Select("cat_name").From("categories")

	cStr, _, err := cQ.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, cStr)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response = make([]string, 0, 20)
	var cat string
	for rows.Next() {
		err = rows.Scan(
			&cat,
		)
		if err != nil {
			continue
		}
		response = append(response, cat)
	}
	return response, nil
}

//returns a list of coupons before the given one and apply a filter if supplied
func (s *Database) CouponBeforeIDAndTimeFiltered(ctx context.Context, couponID string, lastTime string, filter string) (*listing.CouponListResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}
	c := s.psql.Select(`*`).From("categories").Where(sq.Eq{"cat_name": ""}).InnerJoin("coupon_categories using(cat_id")
	d := s.psql.Select("*").FromSelect(c, "cat").InnerJoin("coupons using(coupon_id)").Where(sq.Eq{"coupon_state": "active", "created_at": ""})

	cQ := s.psql.Select(`coupon_id,store_id,store_name,tagline,address,theme_color,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl`).FromSelect(d, "coupons").InnerJoin("stores using(store_id")

	cStr, _, err := cQ.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}

	rows, err := conn.Query(ctx, cStr, filter, lastTime, couponID)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}

	var response listing.CouponListResponse
	for rows.Next() {
		var coupon listing.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Store.ThemeColor,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.TextCouponWebURL,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
		response.LastCouponID = response.Coupons[len(response.Coupons)-1].CouponID
	}

	return &response, nil
}

//returns a list of coupons saved by the user associated with the id
func (s *Database) CouponsSavedByUser(ctx context.Context, userID string) (*listing.CouponListResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}

	c := s.psql.Select(`coupon_id,store_id,store_name,tagline,address,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl`).From("coupons").InnerJoin("saved_coupons using(coupon_id").Where(sq.Eq{"user_id": ""})

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, cStr, userID)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response listing.CouponListResponse
	response.Coupons = make([]listing.Coupon, 0, 10)
	for rows.Next() {
		var coupon listing.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Store.ThemeColor,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.TextCouponWebURL,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
		response.LastCouponID = response.Coupons[len(response.Coupons)-1].CouponID
	}
	if len(response.Coupons) == 0 {
		return nil, listing.ErrCouponsNotFound
	}
	return &response, nil
}

//returns a list of stores followed by a particular user
func (s *Database) StoresFollowedByUser(ctx context.Context, userID string) (*listing.StoreListResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}

	c := s.psql.Select(`store_id,store_name,tagline,address,theme_color`).From("stores").InnerJoin("stores_followed using(store_id").Where("user_id=$ And store_state=active", "")

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, cStr, userID)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}

	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response listing.StoreListResponse
	response.Stores = make([]listing.Store, 0, 10)

	for rows.Next() {
		var store listing.Store
		err = rows.Scan(
			&store.StoreID,
			&store.StoreName,
			&store.Tagline,
			&store.Address,
			&store.ThemeColor,
		)
		if err != nil {
			continue
		}
		response.Stores = append(response.Stores, store)
	}
	if len(response.Stores) == 0 {
		return nil, listing.ErrStoresNotFound
	}

	return &response, nil
}

//returns the store data of the user who owns the store
func (s *Database) UserStoreData(ctx context.Context, storeID string) (*storemanagement.UserStoreResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}

	c := s.psql.Select(`store_id,store_name,tagline,address,theme_color`).From("stores").Where(sq.Eq{"store_id": ""})

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}

	row := conn.QueryRow(ctx, cStr, storeID)

	var response storemanagement.UserStoreResponse

	err = row.Scan(
		&response.Store.StoreID,
		&response.Store.StoreName,
		&response.Store.Tagline,
		&response.Store.Address,
	)
	if err != nil {
		return nil, storemanagement.ErrStoreNotFound
	}

	return &response, nil
}

//returns the list of coupons of the store associated with the user
func (s *Database) UserStoreCoupons(ctx context.Context, storeID string) (*storemanagement.CouponListResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}

	c := s.psql.Select(`coupon_id,store_id,store_name,tagline,address,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl,extract(epoch from created_at`).From("coupons").Where(sq.Eq{"store_id": ""}).OrderBy(" created_at desc")

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, cStr, storeID)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response storemanagement.CouponListResponse

	for rows.Next() {
		var coupon storemanagement.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.CreationTime,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
	}
	return &response, nil
}

// returns the data of the store associated with the storeid
func (s *Database) StoreDetails(ctx context.Context, storeID string) (*listing.SingleStoreResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}

	c := s.psql.Select(`store_id,store_name,tagline,address,theme_color`).From("stores").Where(sq.Eq{"store_id": ""})

	cStr, _, err := c.ToSql()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}

	row := conn.QueryRow(ctx, cStr, storeID)

	var response listing.SingleStoreResponse

	err = row.Scan(
		&response.Store.StoreID,
		&response.Store.StoreName,
		&response.Store.Tagline,
		&response.Store.Address,
		&response.Store.ThemeColor,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, listing.ErrStoreNotFound
		}
		s.logger.Error(err.Error())
		return nil, err
	}

	return &response, nil
}

//returns a list of store coupons of the store associated with the id
func (s *Database) StoreCoupons(ctx context.Context, storeID string) (*listing.CouponListResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}

	c := s.psql.Select(`coupon_id,store_id,store_name,tagline,address,theme_color,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl`).From("coupons").Where("coupon_state=active and store_id =$1").OrderBy("created_at desc")

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}

	rows, err := conn.Query(ctx, cStr, storeID)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response listing.CouponListResponse
	for rows.Next() {
		var coupon listing.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Store.ThemeColor,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.TextCouponWebURL,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
		response.LastCouponID = response.Coupons[len(response.Coupons)-1].CouponID
	}
	if len(response.Coupons) == 0 {
		return nil, listing.ErrCouponsNotFound
	}
	return &response, nil
}

//returns a list of coupons in the category
func (s *Database) CategoryCoupons(ctx context.Context, catName string) (*listing.CouponListResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}

	sub := s.psql.Select(`*`).From("coupon_categories").InnerJoin("categories using(cat_id)").Where(sq.Eq{"cat_name": ""})

	c := s.psql.Select(`coupon_id,store_id,store_name,tagline,address,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl`).FromSelect(sub, "inner").InnerJoin("coupons").Where("coupon_state=active").OrderBy("created_at desc")

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, cStr, catName)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response listing.CouponListResponse
	for rows.Next() {
		var coupon listing.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Store.ThemeColor,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.TextCouponWebURL,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
		response.LastCouponID = response.Coupons[len(response.Coupons)-1].CouponID
	}

	return &response, nil
}

//returns a list of coupons in the category before the given id and created_time
func (s *Database) CategoryCouponsBeforeIDAndTime(ctx context.Context, catName, couponid, lastTime string) (*listing.CouponListResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}

	sub := s.psql.Select(`*`).From("coupon_categories").InnerJoin("categories using(cat_id)").Where(sq.Eq{"cat_name": ""})

	c := s.psql.Select(`coupon_id,store_id,store_name,tagline,address,"desc",amount_off,percentage_off,currency_code,qr_code_url,extract(epoch from expired_at),is_text_coupon,text_coupon_code,text_coupon_weburl`).FromSelect(sub, "inner").InnerJoin("coupons").Where("coupon_state=active").OrderBy("created_at desc")

	cStr, _, err := c.ToSql()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, cStr, catName)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response listing.CouponListResponse
	for rows.Next() {
		var coupon listing.Coupon
		err = rows.Scan(
			&coupon.CouponID,
			&coupon.Store.StoreID,
			&coupon.Store.StoreName,
			&coupon.Store.Tagline,
			&coupon.Store.Address,
			&coupon.Store.ThemeColor,
			&coupon.Desc,
			&coupon.AmountOff,
			&coupon.PercentageOff,
			&coupon.CurrencyCode,
			&coupon.QrCodeURL,
			&coupon.ExpiringDate,
			&coupon.IsTextCoupon,
			&coupon.TextCouponCode,
			&coupon.TextCouponWebURL,
		)
		if err != nil {
			continue
		}
		response.Coupons = append(response.Coupons, coupon)
		response.LastCouponID = response.Coupons[len(response.Coupons)-1].CouponID
	}

	return &response, nil

}

func (s *Database) GetUserStoreCouponsRedeemedCount(ctx context.Context, storeID, days string) (uint, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return 0, storage.ErrServerError
	}

	var c sq.SelectBuilder
	if len(days) == 0 {
		c = s.psql.Select("count(*) from redeemed_coupons").InnerJoin("coupons using(coupon_id)").Where("store_id=$1")
	} else {
		c = s.psql.Select("count(*) from redeemed_coupons").InnerJoin("coupons using(coupon_id)").Where("store_id =$1 and (created_at - now()) <$2", storeID, days)
	}
	cStr, _, err := c.ToSql()
	if err != nil {
		return 0, storage.ErrServerError
	}
	var row pgx.Row
	if len(days) == 0 {
		row = conn.QueryRow(ctx, cStr, storeID)
	}
	row = conn.QueryRow(ctx, cStr, storeID, days)
	var total int
	err = row.Scan(&total)

	if err != nil {
		return 0, err
	}

	return uint(total), nil
}
func (s *Database) CouponState(ctx context.Context, couponid string) (string, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return "", storage.ErrServerError
	}
	c := s.psql.Select("coupon_state").From("coupons").Where(sq.Eq{"coupon_id": ""})
	cStr, _, err := c.ToSql()
	if err != nil {
		return "", storage.ErrServerError
	}
	row := conn.QueryRow(ctx, cStr, couponid)

	var state string
	err = row.Scan(&state)
	if err != nil {
		return "", err
	}
	return state, nil
}

//CreateCoupon create a coupon for the store
func (s *Database) CreateCoupon(ctx context.Context, storeID string, coupon storemanagement.CreateCoupon) (string, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return "", storage.ErrServerError
	}
	created_at := time.Now().UnixNano()

	c := s.psql.Insert("coupons").Columns(`"store_id,"desc",state,discount_typw,created_at,expired_at,percentage_off,amount_off,currency_code,is_text_coupon,text_coupon_weburl,max_redemption,unlimited_redemption`).Values(storeID, coupon.Desc, coupon.State, created_at)

	cStr, _, err := c.ToSql()
	if err != nil {
		return "", storage.ErrServerError
	}

	cStr = cStr + "returning id"
	row := conn.QueryRow(ctx, cStr)
	var couponid string
	err = row.Scan(&couponid)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Message) // => syntax error at end of input
			fmt.Println(pgErr.Code)    // => 42601
		}
		s.logger.Error(err.Error())
		return "", storage.ErrServerError
	}

	return "", nil

}

//DeleteCoupon deletes the coupon from the store
func (s *Database) DeleteCoupon(ctx context.Context, couponID string, storeID string) error {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return storage.ErrServerError
	}

	_, err = conn.Exec(ctx, `delete from coupons where coupon_id = $1 and store_id=$2`, couponID, storeID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Message) // => syntax error at end of input
			fmt.Println(pgErr.Code)    // => 42601
		}
		s.logger.Error(err.Error())
		return storage.ErrServerError
	}

	return nil
}

//Employees returns a list of employees associated with the store
func (s *Database) Employees(ctx context.Context, storeID string) (*storemanagement.EmployeesResponse, error) {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, storage.ErrServerError
	}
	rows, err := conn.Query(ctx, `
	select fullname,emp_id,emp_state from employees inner join users on emp_id=user_id where store_id =$1 AND emp_state <>'removed`, storeID)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, storage.ErrServerError
	}
	var response *storemanagement.EmployeesResponse
	for rows.Next() {
		var employee storemanagement.Employee
		err = rows.Scan(
			&employee.FullName,
			&employee.ID,
			&employee.State,
		)
		if err != nil {
			continue
		}
		response.Employees = append(response.Employees, employee)
	}
	return response, nil

}

//AddEmployee register an employee under the given store
func (s *Database) AddEmployee(ctx context.Context, storeID string, emp storemanagement.Employee) error {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return storage.ErrServerError
	}

	token := storage.GenerateUUID()
	exp := time.Now().UnixNano()
	_, err = conn.Exec(ctx, `insert into create_employees(email,store_id,token,expired_at)values($1,$2,$3`, emp.Email, storeID, token, exp)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil

}

//SuspendEmployee suspends the given  store employee
func (s *Database) SuspendEmployee(ctx context.Context, storeID string, empID string) error {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return storage.ErrServerError
	}
	_, err = conn.Exec(ctx, `
	update stores_employees set emp_state = 'suspended' where emp_id =$1 And store_id =$2`, empID, storeID)
	if err != nil {
		s.logger.Error(err.Error())
		return storage.ErrServerError
	}

	return nil
}

//RemoveEmployee removes the given employee from the store
func (s *Database) RemoveEmployee(ctx context.Context, storeID string, empID string) error {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return storage.ErrServerError
	}

	_, err = conn.Exec(ctx, `
	update stores_employees set emp_state = 'removed' where emp_id =$1 And store_id =$2`, empID, storeID)
	if err != nil {
		s.logger.Error(err.Error())
		return storage.ErrServerError
	}

	return nil
}

//ResumeEmployee resumes the suspended employee
func (s *Database) ResumeEmployee(ctx context.Context, storeID string, empID string) error {
	conn, err := s.dbPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return storage.ErrServerError
	}

	_, err = conn.Exec(ctx, `
	update stores_employees set emp_state = 'active' where emp_id =$1 And store_id =$2`, empID, storeID)
	if err != nil {
		s.logger.Error(err.Error())
		return storage.ErrServerError
	}

	return nil
}

func (s *Database) EditStore(ctx context.Context, userid string, edit storemanagement.StoreEdit) error {
	return nil
}

//VerifyCoupon verifies the coupon
func (s *Database) VerifyCoupon(ctx context.Context, emp_id, couponid string) error {
	return nil
}

//StoreCouponsBeforeIDAndTime returns a list of coupons before the specified id and time
func (s *Database) StoreCouponsBeforeIDAndTime(ctx context.Context, userid, couponid, lastTime string) (*listing.CouponListResponse, error) {
	return nil, nil
}

//QrImage returns the qrimage of the coupon
func (s *Database) QrImage(ctx context.Context, imageID string) ([]byte, error) {
	return storage.GenerateCouponImage("Hello David"), nil
}
