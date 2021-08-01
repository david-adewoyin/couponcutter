package storage

/*
import (
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/tidwall/gjson"
)

func decodeCoupons(couResult gjson.Result) []Coupon {
	coupons := make([]Coupon, 0, 30)
	couResult.ForEach(func(key gjson.Result, value gjson.Result) bool {
		//c, _ := time.Parse(time.RFC3339, value.Get("creation_date").Str)
		//ex, err := time.Parse(time.RFC3339, value.Get("expiring_date").Str)

		//if err != nil {
		//	fmt.Println(err)
		//		os.Exit(1)
		//	}
		co := Coupon{
			CouponID:       value.Get("coupon_id").Str,
			Store:          decodeStore(value.Get("store")),
			AmountOff:      value.Get("offer").Num,
			PercentageOff:  value.Get("offer").Num,
			DiscountType:   value.Get("discount_type").Str,
			Desc:           value.Get("desc").Str,
			Currency:       value.Get("currency").Str,
			CreationDate:   3335046,
			IsTextCoupon:   value.Get("text_only").Bool(),
			ItemURL:        value.Get("item_url").Str,
			TextCouponCode: value.Get("text_coupon_code").Str,
			ExpiringDate:   5455455,
			SingleUserUse:  value.Get("single_use").Bool(),
			MaxRedemption:  MaxRedemption(value.Get("max_redemption").Str),
		}
		coupons = append(coupons, co)

		return true
	})

	return coupons
}
func decodeEmployees(empResult gjson.Result) []Employee {
	employees := make([]Employee, 0, 30)
	empResult.ForEach(func(key gjson.Result, value gjson.Result) bool {

		emp := Employee{
			FullName: value.Get("full_name").Str,
			Email:    value.Get("email").Str,
		}
		employees = append(employees, emp)

		return true
	})
	return employees
}
func decodeStore(value gjson.Result) Store {
	//	value := result.Get("store")

	store := Store{
		StoreID:     value.Get("store_id").Str,
		StoreName:   value.Get("store_name").Str,
		ThemeColor:  int(value.Get("theme_color").Uint()),
		StoreImage:  value.Get("store_image").Str,
		Tagline:     value.Get("tagline").Str,
		Address:     value.Get("address").Str,
		Following:   value.Get("following").Bool(),
		CouponCount: uint(value.Get("coupon_count").Int()),
		Employees:   decodeEmployees(value.Get("employees")),
		Coupons:     decodeCoupons(value.Get("coupons")),
	}
	return store

}
func decodeSubStore(storeResult gjson.Result) []Store {
	subStores := make([]Store, 0, 10)

	storeResult.ForEach(func(key gjson.Result, value gjson.Result) bool {

		store := Store{
			StoreID:     value.Get("store_id").Str,
			StoreName:   value.Get("store_name").Str,
			ThemeColor:  int(value.Get("theme_color").Uint()),
			StoreImage:  value.Get("store_image").Str,
			Tagline:     value.Get("tagline").Str,
			Address:     value.Get("address").Str,
			Following:   value.Get("following").Bool(),
			CouponCount: uint(value.Get("coupon_count").Int()),
			Employees:   decodeEmployees(value.Get("employees")),

			Coupons: decodeCoupons(value.Get("coupons")),
		}
		subStores = append(subStores, store)

		return true
	})
	return subStores
}

//LoadSampleData loads sample data for backend
func (s *Storage) LoadSampleData(coupon io.Reader, store io.Reader) {
	couponJSON, err := ioutil.ReadAll(coupon)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	v := gjson.Get(string(couponJSON), "coupons")
	s.Coupons = decodeCoupons(v)
	storeJSON, err := ioutil.ReadAll(store)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	v = gjson.Get(string(storeJSON), "stores")
	v.ForEach(func(key gjson.Result, value gjson.Result) bool {
		store := Store{
			StoreID:     value.Get("store_id").Str,
			StoreImage:  value.Get("store_image").Str,
			Tagline:     value.Get("tagline").Str,
			Address:     value.Get("address").Str,
			Following:   value.Get("following").Bool(),
			CouponCount: uint(value.Get("coupon_count").Int()),
			Employees:   decodeEmployees(value.Get("employees")),
			SubStores:   decodeSubStore(value.Get("sub_stores")),
			Coupons:     decodeCoupons(value.Get("coupons")),
		}
		s.Stores = append(s.Stores, store)

		return true
	})

}
*/
