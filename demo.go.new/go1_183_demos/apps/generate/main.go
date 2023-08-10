package main

import "fmt"

func main() {
	for i := 0; i <= int(Acetaminophen)+1; i++ {
		p := Pill(i)
		fmt.Println("pill:", p)

		ap := AutoPill(i)
		fmt.Println("auto generate pill:", ap)
	}
}

type Pill int

const (
	Placebo Pill = iota
	Aspirin
	Ibuprofen
	Paracetamol
	Acetaminophen = Paracetamol
)

func (p Pill) String() string {
	switch p {
	case Placebo:
		return "Placebo"
	case Aspirin:
		return "Aspirin"
	case Ibuprofen:
		return "Ibuprofen"
	case Paracetamol:
		return "Paracetamol"
	}
	return fmt.Sprintf("Pill(%d)", p)
}

// AutoPill by generate

//go:generate stringer -type=AutoPill -output=pill_string.go
type AutoPill int

const (
	AutoPlacebo AutoPill = iota
	AutoAspirin
	AutoIbuprofen
	AutoParacetamol
	AutoAcetaminophen = AutoParacetamol
)
