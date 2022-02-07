package storage

type Contest struct {
	Id          uint64 `boltholdKey:"Id"`
	Name        string
	Description string
	When        string
	Where       string
	Closed      bool
	Hidden      bool
}

type ContestParticipant struct {
	Id            uint64 `boltholdKey:"Id"`
	ParticipantId int64
	ContestId     uint64
	Name          string
	School        string
	Contacts      string
	Languages     string
	Login         string
	Password      string
}

type DialogState struct {
	ParticipantId int64 `boltholdKey:"ParticipantId"`
	DialogType    string
	DialogStep    string
	Values        DialogValues
}

type DialogValues map[string]interface{}

type ContestNotification struct {
	Id        uint64 `boltholdKey:"Id"`
	ContestId uint64
	Message   string
}
