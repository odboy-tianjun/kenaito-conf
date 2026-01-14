package service

import (
	"errors"
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("密码错误")
	}

	userInfo := &model.LoginUserInfo{
		Id:              strconv.Itoa(user.Id),
		Name:            user.Name,
		Type:            user.Type,
		GroupId:         user.GroupId,
		GrantAppIds:     make([]int, 0),
		GrantPermissons: make([]string, 0),
	}

	// load permissions
	s.loadPermissions(userInfo, user.Id)

	// generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":  user.Id,
		"name":    user.Name,
		"type":    user.Type,
		"groupId": user.GroupId,
		"exp":     time.Now().Add(time.Duration(s.expireHrs) * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, userInfo, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*model.LoginUserInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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

	userInfo := &model.LoginUserInfo{
		Id:              strconv.Itoa(user.Id),
		Name:            user.Name,
		Type:            user.Type,
		GroupId:         user.GroupId,
		GrantAppIds:     make([]int, 0),
		GrantPermissons: make([]string, 0),
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
		dao.Engine.Where("owner_id = ?", userId).Find(&apps)
	}
	for _, app := range apps {
		userInfo.GrantAppIds = append(userInfo.GrantAppIds, app.Id)
		userInfo.GrantPermissons = append(userInfo.GrantPermissons, buildPermissionCode(app.Id, "RW"))
	}

	// load granted permissions
	var perms []model.UserPermission
	dao.Engine.Where("user_id = ?", userId).Find(&perms)
	for _, p := range perms {
		permCode := buildPermissionCode(p.AppId, p.Operate)
		// avoid duplicate
		found := false
		for _, existing := range userInfo.GrantPermissons {
			if existing == permCode {
				found = true
				break
			}
		}
		if !found {
			userInfo.GrantPermissons = append(userInfo.GrantPermissons, permCode)
		}
		appIdFound := false
		for _, id := range userInfo.GrantAppIds {
			if id == p.AppId {
				appIdFound = true
				break
			}
		}
		if !appIdFound {
			userInfo.GrantAppIds = append(userInfo.GrantAppIds, p.AppId)
		}
	}
}

func buildPermissionCode(appId int, operate string) string {
	return strconv.Itoa(appId) + ":" + operate
}
