package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrServerError = errors.New("the server developed an error while processing the request")
)

//COST for the hashing
const COST = 14

// Coupon represent a redeemable piece of text or qrcode
type Coupon struct {
	CouponID       string   `json:"coupon_id,omitempty"`
	Store          Store    `json:"store,omitempty"`
	AmountOff      float64  `json:"AmountOff,omitempty"`
	PercentageOff  float64  `json:"percentageOff,omitempty"`
	Desc           string   `json:"desc,omitempty"`
	Currency       string   `json:"currency,omitempty"`
	CreationDate   uint     `json:"creation_date,omitempty"`
	ExpiringDate   uint     `json:"expiring_date,omitempty"`
	Categories     []string `json:"categories,omitempty"`
	ThemeColor     int      `json:"theme_color,omitempty"`
	MaxRedemption  string   `json:"max_redemption,omitempty"`
	SingleUserUse  bool     `json:"single_use,omitempty"`
	IsTextCoupon   bool     `json:"text_only,omitempty"`
	TextCouponCode string   `json:"text_coupon_code,omitempty"`
	ItemURL        string   `json:"item_url,omitempty"`
	DiscountType   string   `json:"discount_type,omitempty"`
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

//HashPassword hash the supplied password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), COST)
	if err != nil {
		log.Println(err)
		return "", ErrServerError
	}
	return string(bytes), err
}

// returns a random uuid string
func GenerateUUID() string {
	uuid := uuid.NewString()
	return uuid
}
func CompareUUID(value1, value2 string) bool {
	return value1 == value2

}

//CheckPasswordHash checks if the passwrod and hash are the same
func CheckPasswordHash(hashed, submittedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(submittedPassword))

	return err == nil
}

func createUserWithPassword(email string, password string) error {
	if len(password) > 64 {
		// return password is too long to be accepted
	}
	if len(password) < 5 {
		// return password is too short to be accepted
	}
	return nil

}

//GenerateCouponImage generate the qr code of the coupon
func GenerateCouponImage(desc string) []byte {
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
