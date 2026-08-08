package workflow

type TriggerRaw struct {
	Every   *Cadence     `json:"every"`
	BeginAt *MinuteOfDay `json:"beginAt"`
}
