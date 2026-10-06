package calculator

type Item string

const (
	Red    Item = "red"
	Green  Item = "green"
	Blue   Item = "blue"
	Yellow Item = "yellow"
	Pink   Item = "pink"
	Purple Item = "purple"
	Orange Item = "orange"
)

type Menu map[Item]Money

func DefaultMenu() Menu {
	return Menu{
		Red:    Baht(50),
		Green:  Baht(40),
		Blue:   Baht(30),
		Yellow: Baht(50),
		Pink:   Baht(80),
		Purple: Baht(90),
		Orange: Baht(120),
	}
}
