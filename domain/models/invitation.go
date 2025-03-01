package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Invitation struct {
	ID            bson.ObjectID    `bson:"_id" json:"id"`
	WorkspaceID   bson.ObjectID    `bson:"workspace_id" json:"workspaceId"`
	InviteeUserID bson.ObjectID    `bson:"invitee_user_id" json:"inviteeUserId"`
	Role          InvitationRole   `bson:"role" json:"role"`
	Status        InvitationStatus `bson:"status" json:"status"`
	ExpiredAt     time.Time        `bson:"expired_at" json:"expiredAt"`
	RespondedAt   *time.Time       `bson:"responded_at" json:"respondedAt"`
	CustomMessage *string          `bson:"custom_message" json:"customMessage"`
	CreatedAt     time.Time        `bson:"created_at" json:"createdAt"`
	CreatedBy     bson.ObjectID    `bson:"created_by" json:"createdBy"`
}

type InvitationRole string

const (
	InvitationRoleModerator InvitationRole = "MODERATOR"
	InvitationRoleUser      InvitationRole = "MEMBER"
)

func (i InvitationRole) String() string {
	return string(i)
}

func (i InvitationRole) IsValid() bool {
	switch i {
	case InvitationRoleModerator, InvitationRoleUser:
		return true
	}
	return false
}

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "PENDING"
	InvitationStatusAccepted InvitationStatus = "ACCEPTED"
	InvitationStatusDeclined InvitationStatus = "DECLINED"
	InvitationStatusExpired  InvitationStatus = "EXPIRED"
)

func (i InvitationStatus) String() string {
	return string(i)
}

func (i InvitationStatus) IsValid() bool {
	switch i {
	case InvitationStatusPending, InvitationStatusAccepted, InvitationStatusDeclined, InvitationStatusExpired:
		return true
	}
	return false
}
