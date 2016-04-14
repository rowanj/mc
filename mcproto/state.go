package mcproto

type ConnectionState byte

const (
	Handshaking ConnectionState = iota
	Status
	Login
	Play
)

func (c ConnectionState) String() string {
	if c == Handshaking {
		return "Handshaking"
	}
	if c == Status {
		return "Status"
	}
	if c == Login {
		return "Login"
	}
	if c == Play {
		return "Play"
	}
	return ""
}
