/*
 * Licensed to the AcmeStack under one or more contributor license
 * agreements. See the NOTICE file distributed with this work for
 * additional information regarding copyright ownership.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package openapi

import (
	"fmt"
	"time"

	"github.com/acmestack/envcd/internal/core/storage/dao"
	"github.com/acmestack/envcd/internal/pkg/constant"
	"github.com/acmestack/envcd/internal/pkg/entity"
	"github.com/acmestack/envcd/pkg/entity/result"
	"github.com/acmestack/godkits/array"
	"github.com/acmestack/godkits/gox/stringsx"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// loginParam Login
type loginParam struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// userParam Create User Param
type userParam struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Identity int    `json:"identity"`
	State    string `json:"state"`
}

const (
	// hmacSecret secret
	hmacSecret = "9C035514A15F78"
	userIdKey  = "userId"
	tokenKey   = "accessToken"
)

type pageUserVO struct {
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
	List     []userVO `json:"list"`
}

type userVO struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Identity  int    `json:"identity"`
	State     string `json:"state"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func userConverter(users []entity.User) []userVO {
	var convertUsers []userVO
	if array.Empty(users) {
		return convertUsers
	}
	for _, user := range users {
		convertUsers = append(convertUsers, userVO{
			Id:        user.Id,
			Name:      user.Name,
			Identity:  user.Identity,
			State:     user.State,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return convertUsers
}

// claims claims
type claims struct {
	*jwt.RegisteredClaims
	userId   int
	userName string
}

// newJWTToken secret
func newJWTToken(authClaims claims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims)
	tokenString, err := token.SignedString([]byte(hmacSecret))
	if err != nil {
		// todo
		return ""
	}
	return tokenString
}

func (openapi *Openapi) login(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		param := loginParam{}
		if err := ginCtx.ShouldBindJSON(&param); err != nil {
			// todo log
			//log.Error("Bind error, %v", err)
			return result.InternalFailure(err)
		}

		users, err := dao.New(openapi.storage).SelectUser(entity.User{
			Name: param.Username,
		})
		if err != nil {
			// todo log
			//log.Error("Query User error: %v", err)
			return result.InternalFailure(err)
		}

		if len(users) == 0 {
			// todo log
			//log.Error("User does not exist : %v", param)
			return result.Failure0(result.ErrorUserNotFound)
		}
		user := users[0]
		if saltPassword(param.Password, user.Salt) != user.Password {
			return result.Failure0(result.ErrorUserPasswordIncorrect)
		}
		token := newJWTToken(claims{
			RegisteredClaims: &jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			},
			userId:   user.Id,
			userName: user.Name,
		})
		return result.Success(map[string]interface{}{
			userIdKey: user.Id,
			tokenKey:  token,
		})
	})
}

func (openapi *Openapi) logout(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		fmt.Println("hello world")
		// UserDao.save(),
		// LogDao.save()
		return nil
	})
}

func (openapi *Openapi) createUser(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		param := userParam{}
		if err := ginCtx.ShouldBindJSON(&param); err != nil {
			// todo log
			//log.Error("Bind error, %v", err)
			return result.InternalFailure(err)
		}
		daoAction := dao.New(openapi.storage)
		// check if the user already exists in the database
		users, err := daoAction.SelectUser(entity.User{
			Name: param.Name,
		})
		if err != nil {
			// todo log
			//log.Error("Query User error: %v", err)
			return result.InternalFailure(err)
		}
		if len(users) > 0 {
			// todo log
			//log.Error("User Has exists: %v", users)
			return result.Failure0(result.ErrorUserExisted)
		}
		// generate database password by salt
		salt := randomSalt()
		password := saltPassword(param.Password, salt)
		user := entity.User{
			Name:      param.Name,
			Password:  password,
			Salt:      salt,
			Identity:  param.Identity,
			State:     stringsx.DefaultIfEmpty(param.State, constant.EnabledState),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// save user
		if _, _, err := daoAction.InsertUser(user); err != nil {
			// todo log
			//log.Error("insert error=%v", err)
			return result.Failure(result.ErrorCreateUser, err)
		}
		// fixme update success message or response token and id ?
		return result.Success("ok")
	})
}

func (openapi *Openapi) updateUser(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		fmt.Println("hello world")
		return nil
	})
}

func (openapi *Openapi) user(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		id := stringsx.ToInt(ginCtx.Param("userId"))
		param := entity.User{Id: id}
		// query user by param
		users, err := dao.New(openapi.storage).SelectUser(param)
		if err != nil {
			// todo log
			//log.Error("select user error = %v", err)
			return result.Failure(result.ErrorUserNotFound, err)
		}
		if len(users) == 0 {
			// todo log
			//log.Error("User does not exist : %v", param)
			return result.Failure0(result.ErrorUserNotFound)
		}
		return result.Success(userVO{
			Id:        users[0].Id,
			Name:      users[0].Name,
			Identity:  users[0].Identity,
			State:     users[0].State,
			CreatedAt: users[0].CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: users[0].UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	})
}

func (openapi *Openapi) removeUser(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		id := stringsx.ToInt(ginCtx.Param("userId"))
		param := entity.User{Id: id}

		daoAction := dao.New(openapi.storage)
		user, err := daoAction.SelectUserById(param)
		if err != nil {
			// todo log
			//log.Error("select users error = %v", err)
			return result.Failure(result.ErrorUserNotFound, err)
		}
		if user.Id == 0 {
			return result.Success(nil)
		}
		// update user state to deleted
		// todo tx and catch error @liuzhaowei
		user.State = constant.DeletedState
		daoAction.UpdateUser(user)

		// get all the user's dictionary and update state to deleted
		dictParam := entity.Dictionary{UserId: user.Id}
		dictionaries, err := daoAction.SelectDictionary(dictParam, nil)
		if err != nil {
			return result.InternalFailure(err)
		}
		if len(dictionaries) != 0 {
			for i := range dictionaries {
				dictionaries[i].State = constant.DeletedState
			}
			daoAction.UpdateDictionaryBatch(dictionaries)
		}

		// get all the user's scopespace and update state to deleted
		spaceParam := entity.ScopeSpace{UserId: user.Id}
		spaces, err := daoAction.SelectScopeSpace(spaceParam)
		if err != nil {
			return result.InternalFailure(err)
		}
		if len(spaces) != 0 {
			for i := range spaces {
				spaces[i].State = constant.DeletedState
			}
			daoAction.UpdateScopeSpaceBatch(spaces)
		}

		return result.Success(nil)
	})
}

func (openapi *Openapi) users(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		// receive params from request
		// todo use ToIntDefault func
		page := stringsx.ToInt(ginCtx.Query("page"))
		pageSize := stringsx.ToIntOrDefault(ginCtx.Query("pageSize"), 20)
		if page == 0 {
			page = 1
		}
		nameParam := ginCtx.Query("name")

		pageParam := entity.PageUserParam{Page: page, PageSize: pageSize, Name: nameParam}

		users, err := dao.New(openapi.storage).PageSelectUser(pageParam)
		if err != nil {
			// todo log
			//log.Error("select users error = %v", err)
			return result.Failure(result.ErrorUserNotFound, err)
		}
		// todo use PageListVO
		return result.Success(pageUserVO{
			page, pageSize, userConverter(users),
		})
	})
}

func (openapi *Openapi) userScopeSpaces(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		fmt.Println("hello world")
		return nil
	})
}

func (openapi *Openapi) userDictionaries(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		fmt.Println("hello world")
		return nil
	})
}

func (openapi *Openapi) userDictionariesUnderScopeSpace(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		fmt.Println("hello world")
		return nil
	})
}
