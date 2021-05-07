func loadSampleData(storage *memory.Storage) {
	coupon, err := os.Open("coupon.json")
	if err != nil {
		os.Exit(1)
	}
	store, err := os.Open("store.json")
	if err != nil {
		os.Exit(1)
	}

	_ = coupon
	_ = store
	//storage.LoadSampleData(coupon, store)

}