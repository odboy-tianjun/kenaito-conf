package service

import (
	"errors"
	"fmt"
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/model"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwtSecret string
	expireHrs int
}

func NewAuthService(secret string, expireHours int) *AuthService {
	return &AuthService{jwtSecret: secret, expireHrs: expireHours}
}

func (s *AuthService) Login(userName, password string) (string, *model.LoginUserInfo, error) {
	var user model.User
	has, err := dao.Engine.Where("name = ? AND enabled = 1", userName).Get(&user)
	if err != nil {
		return "", nil, err
	}
	if !has {
		return "", nil, errors.New("用户不存在或已禁用")
	}

	// 检查业务组是否启用
	if user.GroupId != nil {
		var group model.BusinessGroup
		has, _ := dao.Engine.ID(*user.GroupId).Get(&group)
		if !has || !group.Enabled {
			return "", nil, errors.New("所属业务组已禁用")
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("密码错误")
	}

	userInfo := &model.LoginUserInfo{
		Id:          strconv.Itoa(user.Id),
		Name:        user.Name,
		Type:        user.Type,
		GroupId:     user.GroupId,
		GrantAppIds: make([]int, 0),
	}

	// load permissions
	s.loadPermissions(userInfo, user.Id)

	// generate JWT
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":  user.Id,
		"name":    user.Name,
		"type":    user.Type,
		"groupId": user.GroupId,
		"iat":     now.Unix(),
		"exp":     now.Add(time.Duration(s.expireHrs) * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, userInfo, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*model.LoginUserInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	userId := int(claims["userId"].(float64))
	var user model.User
	has, err := dao.Engine.ID(userId).Get(&user)
	if err != nil || !has {
		return nil, errors.New("user not found")
	}

	// 检查用户是否已被禁用
	if !user.Enabled {
		return nil, errors.New("user is disabled")
	}

	// 检查业务组是否启用
	if user.GroupId != nil {
		var group model.BusinessGroup
		has, _ := dao.Engine.ID(*user.GroupId).Get(&group)
		if !has || !group.Enabled {
			return nil, errors.New("group is disabled")
		}
	}

	// 检查 token 是否在登出后签发
	if user.LastLogoutAt != nil {
		iat := int64(claims["iat"].(float64))
		if iat <= user.LastLogoutAt.Unix() {
			return nil, errors.New("token is invalid after logout")
		}
	}

	userInfo := &model.LoginUserInfo{
		Id:          strconv.Itoa(user.Id),
		Name:        user.Name,
		Type:        user.Type,
		GroupId:     user.GroupId,
		GrantAppIds: make([]int, 0),
	}
	s.loadPermissions(userInfo, user.Id)
	return userInfo, nil
}

func (s *AuthService) loadPermissions(userInfo *model.LoginUserInfo, userId int) {
	if userInfo.IsSuperAdmin() {
		return
	}

	// load owned apps
	var apps []model.Application
	if userInfo.IsGroupAdmin() {
		dao.Engine.Where("group_id = ?", userInfo.GroupId).Find(&apps)
	} else {
		userIdStr := strconv.Itoa(userId)
		dao.Engine.Where("FIND_IN_SET(?, owner_ids) > 0", userIdStr).Find(&apps)
	}
	for _, app := range apps {
		userInfo.GrantAppIds = append(userInfo.GrantAppIds, app.Id)
	}
}
