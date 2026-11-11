package notification

import (
	"testing"
	"time"

	"simplejourney/helpdesk/backend/internal/domain"
)

func TestHub_BroadcastAndTargeting(t *testing.T) {
	hub := NewHub()

	agentClient := &Client{
		UserID: 10,
		Role:   domain.RoleAgent,
		Send:   make(chan Event, 8),
	}
	requesterClient := &Client{
		UserID: 20,
		Role:   domain.RoleRequester,
		Send:   make(chan Event, 8),
	}

	hub.Register(agentClient)
	hub.Register(requesterClient)

	// 1. Broadcast to agent and admin
	createEvt := Event{
		Type:      "ticket_created",
		TicketID:  1,
		Title:     "Kendala Jaringan",
		CreatedAt: time.Now(),
	}
	hub.BroadcastToRoles(createEvt, domain.RoleAgent, domain.RoleAdmin)

	select {
	case received := <-agentClient.Send:
		if received.TicketID != 1 {
			t.Errorf("expected ticket ID 1, got %d", received.TicketID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("agent client did not receive broadcast")
	}

	select {
	case <-requesterClient.Send:
		t.Errorf("requester should not have received ticket_created broadcast")
	default:
	}

	// 2. Direct message to specific user
	directEvt := Event{
		Type:      "ticket_status_updated",
		TicketID:  1,
		Title:     "Status Resolved",
		CreatedAt: time.Now(),
	}
	hub.SendToUsers(directEvt, 20)

	select {
	case received := <-requesterClient.Send:
		if received.Type != "ticket_status_updated" {
			t.Errorf("expected ticket_status_updated, got %s", received.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("requester client did not receive targeted event")
	}

	select {
	case <-agentClient.Send:
		t.Errorf("agent client should not have received targeted event for user 20")
	default:
	}

	hub.Unregister(agentClient)
	hub.Unregister(requesterClient)
}
