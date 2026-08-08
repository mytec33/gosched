package workflow

// TriggerRaw is solely used to store the decoded data from Unmarshal()
// and will be immediately converted to non-pointer form in type Workflow
type TriggerRaw struct {
	Every   *Cadence     `json:"every"`
	BeginAt *MinuteOfDay `json:"beginAt"`
}
