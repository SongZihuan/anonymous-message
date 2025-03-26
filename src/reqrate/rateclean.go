package reqrate

func RateClean() error {
	go CleanAddressRate()
	go CleanIPRate()
	go CleanUserEmailRate()

	return nil
}
