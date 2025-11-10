package panicpkg

func trigger() {
	panic("boom") // want "использование встроенной функции panic запрещено"
}
