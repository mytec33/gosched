package types

type CadenceMeasure struct {
	v string
}

func CadenceDay() CadenceMeasure {
	return CadenceMeasure{v: "d"}
}

func CadenceHour() CadenceMeasure {
	return CadenceMeasure{v: "h"}
}

func CadenceMinute() CadenceMeasure {
	return CadenceMeasure{v: "m"}
}

func (m CadenceMeasure) String() string {
	return m.v
}
