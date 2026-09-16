package service

import (
	"testing"

	"kenaito-conf/server/internal/model"
)

func TestRequirePermission(t *testing.T) {
	svc := NewPermissionService()
	groupID := 1
	otherGroupID := 2
	appID := 10

	tests := []struct {
		name    string
		user    *model.LoginUserInfo
		groupId *int
		appId   *int
		wantErr bool
	}{
		{
			name:    "super admin can access anything",
			user:    &model.LoginUserInfo{Type: "superAdmin"},
			groupId: &groupID,
			appId:   &appID,
		},
		{
			name:    "group admin can access own group",
			user:    &model.LoginUserInfo{Type: "groupAdmin", GroupId: &groupID},
			groupId: &groupID,
			appId:   &appID,
		},
		{
			name:    "group admin cannot access other group",
			user:    &model.LoginUserInfo{Type: "groupAdmin", GroupId: &groupID},
			groupId: &otherGroupID,
			appId:   &appID,
			wantErr: true,
		},
		{
			name:    "owner can access app",
			user:    &model.LoginUserInfo{Type: "user", GroupId: &groupID, GrantAppIds: []int{appID}},
			groupId: &groupID,
			appId:   &appID,
		},
		{
			name:    "owner can access app across group",
			user:    &model.LoginUserInfo{Type: "user", GroupId: &otherGroupID, GrantAppIds: []int{appID}},
			groupId: &groupID,
			appId:   &appID,
		},
		{
			name:    "normal group member cannot access app without ownership",
			user:    &model.LoginUserInfo{Type: "user", GroupId: &groupID},
			groupId: &groupID,
			appId:   &appID,
			wantErr: true,
		},
		{
			name:    "normal user cannot access global config",
			user:    &model.LoginUserInfo{Type: "user", GroupId: &groupID},
			groupId: &groupID,
			appId:   nil,
			wantErr: true,
		},
		{
			name:    "group admin can access own global config",
			user:    &model.LoginUserInfo{Type: "groupAdmin", GroupId: &groupID},
			groupId: &groupID,
			appId:   nil,
		},
		{
			name:    "group admin cannot access other global config",
			user:    &model.LoginUserInfo{Type: "groupAdmin", GroupId: &groupID},
			groupId: &otherGroupID,
			appId:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.RequirePermission(tt.user, tt.groupId, tt.appId)
			if tt.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
