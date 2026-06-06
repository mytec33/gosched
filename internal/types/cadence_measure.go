package types

type CadenceMeasure struct {
	v string
}

var (
	CadenceDay    = CadenceMeasure{"d"}
	CadenceHour   = CadenceMeasure{"h"}
	CadenceMinute = CadenceMeasure{"m"}
)

func (m CadenceMeasure) String() string {
	return m.v
}
