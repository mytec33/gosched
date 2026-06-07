package types

type CadenceMeasure struct {
	v string
}

var (
	CadenceDay    = CadenceMeasure{v: "d"}
	CadenceHour   = CadenceMeasure{v: "h"}
	CadenceMinute = CadenceMeasure{v: "m"}
)

func (m CadenceMeasure) String() string {
	return m.v
}
