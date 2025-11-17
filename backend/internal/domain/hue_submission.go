package domain

type HueResultSubmission struct {
	userName Name
	record   HueRecord
}

func NewHueResultSubmission(userName Name, record HueRecord) HueResultSubmission {
	return HueResultSubmission{
		userName: userName,
		record:   record,
	}
}

func (s HueResultSubmission) UserName() Name {
	return s.UserName()
}

func (s HueResultSubmission) Record() HueRecord {
	return s.record
}
