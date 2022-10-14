package rest

type Error struct {
	Message string
	Class   int
}

func (eS Error) Error() string {
	return eS.Message
}

func MakeError(err string, class int) error {
	return &Error{Message: err, Class: class}
}
