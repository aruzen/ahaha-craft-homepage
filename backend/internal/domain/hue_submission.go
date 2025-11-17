package domain

type HueResultSubmission struct {
	session SessionData
	record  HueRecord
}

func NewHueResultSubmission(session SessionData, record HueRecord) HueResultSubmission {
	return HueResultSubmission{
		session: session,
		record:  record,
	}
}

func (s HueResultSubmission) Session() SessionData {
	return s.session
}

func (s HueResultSubmission) Record() HueRecord {
	return s.record
}
