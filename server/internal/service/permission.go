package service

import (
	"errors"
	"kenaito-conf/server/internal/model"
)

type PermissionService struct{}

func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

func (s *PermissionService) RequireSuperAdmin(userInfo *model.LoginUserInfo) error {
	if !userInfo.IsSuperAdmin() {
		return errors.New("超级管理员才有权限操作")
	}
	return nil
}

func (s *PermissionService) RequirePermission(userInfo *model.LoginUserInfo, groupId *int, appId *int) error {
	if userInfo.IsSuperAdmin() {
		return nil
	}
	if userInfo.IsGroupAdmin() && groupId != nil && userInfo.GroupId != nil && *groupId == *userInfo.GroupId {
		return nil
	}
	if appId == nil || *appId == 0 {
		return errors.New("你没有全局配置权限")
	}
	for _, id := range userInfo.GrantAppIds {
		if id == *appId {
			return nil
		}
	}
	return errors.New("你没有该应用权限")
}

func (s *PermissionService) HasPermissionForUser(userInfo *model.LoginUserInfo, targetUser *model.User) error {
	if userInfo.IsSuperAdmin() {
		return nil
	}
	if !userInfo.IsGroupAdmin() {
		return errors.New("你不是业务组管理员")
	}
	if userInfo.GroupId == nil || targetUser.GroupId == nil || *userInfo.GroupId != *targetUser.GroupId {
		return errors.New("不能操作其他组成员")
	}
	return nil
}
