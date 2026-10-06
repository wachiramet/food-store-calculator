package calculator

import "fmt"

type Money int64

const satangPerBaht = 100

func Baht(amount int64) Money {
	return Money(amount * satangPerBaht)
}

func (m Money) Percent(percent int64) Money {
	return m * Money(percent) / 100
}

func (m Money) String() string {
	sign := ""
	if m < 0 {
		sign = "-"
		m = -m
	}
	return fmt.Sprintf("%s%d.%02d", sign, m/satangPerBaht, m%satangPerBaht)
}
