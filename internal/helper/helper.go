package helper

// Assert panic if err != nil
func Assert(fun func() error) {
	if err := fun(); err != nil {
		panic(err)
	}
}
