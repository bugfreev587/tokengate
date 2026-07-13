package middleware

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	byoSubscriptionRequiredMessage  = "BYO subscription required. Subscribe to enable this connected account."
	byoConnectedAccountErrorMessage = "Connected BYO account is unavailable. Check or reconnect the provider account."
)

type byoAdmissionError struct {
	Status  int
	Code    string
	Message string
}

func evaluateBYOAdmission(apiKey *service.APIKey) *byoAdmissionError {
	if apiKey == nil ||
		!service.IsUserOwnedConnectedAccountCapacity(apiKey.User, apiKey.Group) ||
		apiKey.Group.BYOEnabled == nil ||
		*apiKey.Group.BYOEnabled {
		return nil
	}

	switch apiKey.Group.BYODisabledReason {
	case service.BYOAccountDisabledReasonSubscriptionInactive:
		return &byoAdmissionError{Status: http.StatusForbidden, Code: "SUBSCRIPTION_REQUIRED", Message: byoSubscriptionRequiredMessage}
	case service.BYOAccountDisabledReasonAccountDisabled, service.BYOAccountDisabledReasonNoAccount:
		return &byoAdmissionError{Status: http.StatusServiceUnavailable, Code: "CONNECTED_ACCOUNT_ERROR", Message: byoConnectedAccountErrorMessage}
	default:
		return nil
	}
}
