package workflow

type CadenceMeasure struct {
	v string
}

var (
	CadenceDay     = CadenceMeasure{v: "d"}
	CadenceHour    = CadenceMeasure{v: "h"}
	CadenceMinute  = CadenceMeasure{v: "m"}
	CadenceUnknown = CadenceMeasure{v: "unknown"}
)

func (m CadenceMeasure) String() string {
	return m.v
}
