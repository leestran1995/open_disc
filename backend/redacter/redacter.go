package redacter

import (
	"backend/model"
)

func RedactServerEvent(serverEvent model.ServerEvent, userRoles []string) *model.ServerEvent {
	userRoleMap := make(map[string]bool)
	for _, role := range userRoles {
		userRoleMap[role] = true
	}

	for _, role := range *serverEvent.RoleScope {
		if userRoleMap[role] {
			return &serverEvent
		}
	}

	return RedactedServerEvent(serverEvent)
}

func RedactedServerEvent(serverEvent model.ServerEvent) *model.ServerEvent {
	return &model.ServerEvent{
		ServerEventType:  model.Redacted,
		ServerEventOrder: serverEvent.ServerEventOrder,
		ServerEventTime:  serverEvent.ServerEventTime,
		Payload:          nil,
	}
}
