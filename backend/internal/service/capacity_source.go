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

func isUserOwnedConnectedAccountCapacity(user *User, group *Group) bool {
	if user == nil || group == nil || group.OwnerUserID == nil {
		return false
	}
	return *group.OwnerUserID == user.ID &&
		NormalizeCapacitySource(group.CapacitySource) == CapacitySourceConnectedAccount
}

func IsUserOwnedConnectedAccountCapacity(user *User, group *Group) bool {
	return isUserOwnedConnectedAccountCapacity(user, group)
}

func resolveCapacitySourceForUserGroup(user *User, group *Group) string {
	if isUserOwnedConnectedAccountCapacity(user, group) {
		return CapacitySourceConnectedAccount
	}
	return CapacitySourceTokenGate
}

func shouldChargeTokenGateCapacity(capacitySource string) bool {
	return NormalizeCapacitySource(capacitySource) != CapacitySourceConnectedAccount
}

func zeroTokenGateActualCost(capacitySource string, cost *CostBreakdown) *CostBreakdown {
	if shouldChargeTokenGateCapacity(capacitySource) || cost == nil {
		return cost
	}
	zeroed := *cost
	zeroed.ActualCost = 0
	return &zeroed
}
