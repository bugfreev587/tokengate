package service

const (
	CapacitySourceTokenGate        = "tokengate"
	CapacitySourceConnectedAccount = "connected_account"
)

func NormalizeCapacitySource(value string) string {
	switch value {
	case CapacitySourceConnectedAccount:
		return CapacitySourceConnectedAccount
	default:
		return CapacitySourceTokenGate
	}
}
