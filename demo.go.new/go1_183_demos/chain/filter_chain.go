package chain

type Person struct {
	Name  string
	Age   int
	Title string
}

type Filter interface {
	DoFilter(persons []Person) []Person
}

type AgeFilter struct {
	AgeCond int
}

func (f AgeFilter) DoFilter(persons []Person) []Person {
	results := make([]Person, len(persons)/2)
	for _, p := range persons {
		if p.Age >= f.AgeCond {
			results = append(results, p)
		}
	}
	return results
}

type TitleFilter struct {
	TitleCond string
}

func (f TitleFilter) DoFilter(persons []Person) []Person {
	results := make([]Person, len(persons)/2)
	for _, p := range persons {
		if p.Title == f.TitleCond {
			results = append(results, p)
		}
	}
	return results
}

func RunFilter(persons []Person, filters []Filter) []Person {
	for _, filter := range filters {
		persons = filter.DoFilter(persons)
	}
	return persons
}
