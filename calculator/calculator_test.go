package calculator

import (
	"errors"
	"reflect"
	"testing"
)

func newDefault(t *testing.T) *Calculator {
	t.Helper()

	calc, err := NewDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return calc
}

func TestPrice(t *testing.T) {
	tests := []struct {
		name  string
		order Order
		want  Money
	}{
		{
			name:  "empty order",
			order: Order{},
			want:  0,
		},
		{
			name:  "red and green",
			order: Order{Items: map[Item]int{Red: 1, Green: 1}},
			want:  Baht(90),
		},
		{
			name:  "red and green with member card",
			order: Order{Items: map[Item]int{Red: 1, Green: 1}, HasMemberCard: true},
			want:  Baht(81),
		},
		{
			name: "one of each",
			order: Order{Items: map[Item]int{
				Red: 1, Green: 1, Blue: 1, Yellow: 1, Pink: 1, Purple: 1, Orange: 1,
			}},
			want: Baht(460),
		},
		{
			name:  "zero quantity is ignored",
			order: Order{Items: map[Item]int{Red: 1, Blue: 0}},
			want:  Baht(50),
		},
		{
			name:  "pair of orange",
			order: Order{Items: map[Item]int{Orange: 2}},
			want:  Baht(228),
		},
		{
			name:  "pair of pink",
			order: Order{Items: map[Item]int{Pink: 2}},
			want:  Baht(152),
		},
		{
			name:  "pair of green",
			order: Order{Items: map[Item]int{Green: 2}},
			want:  Baht(76),
		},
		{
			name:  "three green only discounts one pair",
			order: Order{Items: map[Item]int{Green: 3}},
			want:  Baht(116),
		},
		{
			name:  "five orange discounts two pairs",
			order: Order{Items: map[Item]int{Orange: 5}},
			want:  Baht(576),
		},
		{
			name:  "pair of red has no discount",
			order: Order{Items: map[Item]int{Red: 2}},
			want:  Baht(100),
		},
		{
			name:  "orange and pink do not make a pair",
			order: Order{Items: map[Item]int{Orange: 1, Pink: 1}},
			want:  Baht(200),
		},
		{
			name:  "pairs of orange, pink and green",
			order: Order{Items: map[Item]int{Orange: 2, Pink: 2, Green: 2}},
			want:  Baht(456),
		},
		{
			name:  "pair discount does not touch other items",
			order: Order{Items: map[Item]int{Orange: 2, Purple: 1}},
			want:  Baht(318),
		},
		{
			name:  "member discount applies after pair discount",
			order: Order{Items: map[Item]int{Orange: 2}, HasMemberCard: true},
			want:  Money(20520),
		},
		{
			name:  "five orange and red with member card",
			order: Order{Items: map[Item]int{Orange: 5, Red: 1}, HasMemberCard: true},
			want:  Money(56340),
		},
	}

	calc := newDefault(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Price(tt.order)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestReceipt(t *testing.T) {
	tests := []struct {
		name  string
		order Order
		want  Receipt
	}{
		{
			name:  "no discount",
			order: Order{Items: map[Item]int{Red: 1, Green: 1}},
			want:  Receipt{Subtotal: Baht(90), Total: Baht(90)},
		},
		{
			name:  "bundle discount only",
			order: Order{Items: map[Item]int{Orange: 2}},
			want: Receipt{
				Subtotal:  Baht(240),
				Discounts: []AppliedDiscount{{Name: "Bundle orange x1 (5%)", Amount: Baht(12)}},
				Total:     Baht(228),
			},
		},
		{
			name:  "member discount only",
			order: Order{Items: map[Item]int{Red: 1, Green: 1}, HasMemberCard: true},
			want: Receipt{
				Subtotal:  Baht(90),
				Discounts: []AppliedDiscount{{Name: "Member card (10%)", Amount: Baht(9)}},
				Total:     Baht(81),
			},
		},
		{
			name:  "one bundle line per item, ordered by item, then member discount",
			order: Order{Items: map[Item]int{Orange: 5, Green: 3, Red: 1}, HasMemberCard: true},
			want: Receipt{
				Subtotal: Baht(770),
				Discounts: []AppliedDiscount{
					{Name: "Bundle green x1 (5%)", Amount: Baht(4)},
					{Name: "Bundle orange x2 (5%)", Amount: Baht(24)},
					{Name: "Member card (10%)", Amount: Money(7420)},
				},
				Total: Money(66780),
			},
		},
		{
			name:  "bundle item below bundle size has no line",
			order: Order{Items: map[Item]int{Orange: 2, Pink: 1}},
			want: Receipt{
				Subtotal:  Baht(320),
				Discounts: []AppliedDiscount{{Name: "Bundle orange x1 (5%)", Amount: Baht(12)}},
				Total:     Baht(308),
			},
		},
		{
			name:  "member card on empty order",
			order: Order{HasMemberCard: true},
			want:  Receipt{},
		},
	}

	calc := newDefault(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Receipt(tt.order)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPriceErrors(t *testing.T) {
	tests := []struct {
		name  string
		order Order
		want  error
	}{
		{
			name:  "unknown item",
			order: Order{Items: map[Item]int{"black": 1}},
			want:  ErrUnknownItem,
		},
		{
			name:  "negative quantity",
			order: Order{Items: map[Item]int{Red: -1}},
			want:  ErrInvalidQuantity,
		},
	}

	calc := newDefault(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := calc.Price(tt.order)
			if !errors.Is(err, tt.want) {
				t.Errorf("got %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCustomMenuAndDiscounts(t *testing.T) {
	menu := Menu{"black": Baht(200)}
	calc, err := New(
		menu,
		BundleDiscount{Items: []Item{"black"}, Size: 3, Percent: 10},
		MemberDiscount{Percent: 20},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := calc.Price(Order{Items: map[Item]int{"black": 4}, HasMemberCard: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Baht(592)
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestNewRejectsInvalidDiscounts(t *testing.T) {
	tests := []struct {
		name     string
		discount Discount
	}{
		{
			name:     "bundle size zero",
			discount: BundleDiscount{Items: []Item{Orange}, Size: 0, Percent: 5},
		},
		{
			name:     "bundle size negative",
			discount: BundleDiscount{Items: []Item{Orange}, Size: -1, Percent: 5},
		},
		{
			name:     "bundle percent negative",
			discount: BundleDiscount{Items: []Item{Orange}, Size: 2, Percent: -1},
		},
		{
			name:     "bundle percent above 100",
			discount: BundleDiscount{Items: []Item{Orange}, Size: 2, Percent: 101},
		},
		{
			name:     "member percent negative",
			discount: MemberDiscount{Percent: -1},
		},
		{
			name:     "member percent above 100",
			discount: MemberDiscount{Percent: 101},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(DefaultMenu(), tt.discount)
			if !errors.Is(err, ErrInvalidDiscount) {
				t.Errorf("got %v, want %v", err, ErrInvalidDiscount)
			}
		})
	}
}

func TestNoDiscounts(t *testing.T) {
	calc, err := New(DefaultMenu())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := calc.Price(Order{Items: map[Item]int{Orange: 2}, HasMemberCard: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Baht(240)
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestDiscountsRunInGivenOrder(t *testing.T) {
	calc, err := New(
		DefaultMenu(),
		MemberDiscount{Percent: 10},
		BundleDiscount{Items: []Item{Orange}, Size: 2, Percent: 5},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := calc.Price(Order{Items: map[Item]int{Orange: 2}, HasMemberCard: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Baht(204)
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestMoneyPercent(t *testing.T) {
	tests := []struct {
		money   Money
		percent int64
		want    Money
	}{
		{Baht(240), 5, Baht(12)},
		{Baht(626), 10, Money(6260)},
		{Money(3333), 10, Money(333)},
		{Money(99), 5, Money(4)},
	}

	for _, tt := range tests {
		if got := tt.money.Percent(tt.percent); got != tt.want {
			t.Errorf("%s at %d%%: got %s, want %s", tt.money, tt.percent, got, tt.want)
		}
	}
}

func TestMoneyString(t *testing.T) {
	tests := []struct {
		money Money
		want  string
	}{
		{Baht(0), "0.00"},
		{Baht(90), "90.00"},
		{Money(20520), "205.20"},
		{Money(5), "0.05"},
		{Money(-5), "-0.05"},
		{Money(-12345), "-123.45"},
	}

	for _, tt := range tests {
		if got := tt.money.String(); got != tt.want {
			t.Errorf("got %s, want %s", got, tt.want)
		}
	}
}
