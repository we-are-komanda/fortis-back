package rdbms

type TXIsoLevel int8

const (
	TXIsoLevelDefault TXIsoLevel = iota
	TXIsoLevelReadCommitted
	TXIsoLevelRepeatableRead
	TXIsoLevelSerializable
)

func (l TXIsoLevel) String() string {
	switch l {
	case TXIsoLevelReadCommitted:
		return "READ COMMITTED"
	case TXIsoLevelRepeatableRead:
		return "REPEATABLE READ"
	case TXIsoLevelSerializable:
		return "SERIALIZABLE"
	default:
		return ""
	}
}
