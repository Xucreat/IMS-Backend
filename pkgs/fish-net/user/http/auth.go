package http

import (
	"encoding/binary"
	"fishnet/domain"
	"fishnet/glb"
	"fishnet/user/usecase"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"go.uber.org/zap"
)

var _userUsecase domain.UserUsecase

func init() {
	_userUsecase = usecase.NewUserUsecase()
}

const SESSION_DATA_KEY = "registeration"

func ByteToInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}

func Play(c *gin.Context) {
	session := sessions.Default(c)
	var count int
	v := session.Get("count")
	if v == nil {
		count = 0
	} else {
		count = v.(int)
		count++
	}
	session.Set("count", count)
	session.Save()
	c.JSON(200, gin.H{"count": count})
}

func RegisterBegin(c *gin.Context) {
	userName, ok := c.Params.Get("username")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "muse give a username",
		})
		return
	}
	glb.LOG.Info("RegisterBegin username: " + userName)

	// 没有用户就新建
	users, count, err := _userUsecase.QueryUser(&userName, 1, 0)
	if err != nil || count == 0 {
		glb.LOG.Info("Creating user: " + userName)
		user := &domain.User{
			Username: userName,
			Nickname: userName,
			Icon:     "https://pics.com/avatar.png",
		}
		err = _userUsecase.CreateUser([]*domain.User{user})
		if err != nil {
			glb.LOG.Warn("USER CREATION ERROR: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"failed": true,
				"msg":    "user not created",
			})
			return
		}
	} else {
		glb.LOG.Info("user exists")
	}

	users, count, err = _userUsecase.QueryUser(&userName, 1, 0)
	user := users[0]

	// 检查用户已经注册的设备，不允许重复注册
	registerOptions := func(credCreationOpts *protocol.PublicKeyCredentialCreationOptions) {
		credCreationOpts.CredentialExcludeList = user.CredentialExcludeList()
	}
	options, sessionData, err := glb.Auth.BeginRegistration(user, registerOptions)
	if err != nil {
		glb.LOG.Warn("SESSION ERROR: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"failed": true,
			"msg":    "server is broken",
		})
		return
	}

	// store the sessionData values
	session := sessions.Default(c)
	session.Set(SESSION_DATA_KEY, *sessionData)
	session.Set("a", 1)
	err = session.Save()
	if err != nil {
		msg := "can not save session"
		glb.LOG.Warn(msg + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"failed": true,
			"msg":    msg,
		})
		return
	}

	// return the options generated
	c.JSON(http.StatusOK, gin.H{
		"options": options,
		"user": map[string]any{
			"id": user.ID,
		},
	})
	// options.publicKey contain our registration options
}

func RegisterFinish(c *gin.Context) {

	idString, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "must be registering",
		})
		return
	}
	idUint, err := strconv.ParseInt(idString, 10, 64)
	glb.LOG.Info("field to query", zap.Int64("idUint", idUint))
	users, err := _userUsecase.MGetUsers([]int64{idUint})
	if err != nil || len(users) == 0 {
		glb.LOG.Warn("user not exists" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "user not exists " + idString,
		})
		return
	}
	user := users[0]

	// Get the session data stored from the function above
	// using gorilla/sessions it could look like this
	session := sessions.Default(c)

	sessionData := session.Get(SESSION_DATA_KEY)
	if sessionData == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "get sessionData failed",
		})
		return
	}
	glb.LOG.Info("RegisterBegin sessionData.UserID: " + sessionData.(webauthn.SessionData).Challenge)

	response, err := protocol.ParseCredentialCreationResponseBody(c.Request.Body)
	glb.LOG.Info("\nRegisterBegin request origin: " + response.Response.CollectedClientData.Origin + "\n")
	glb.LOG.Info("\nRegisterBegin allowed origin: " + strings.Join(glb.Auth.Config.RPOrigins, "|") + "\n")

	if err != nil {
		glb.LOG.Info("FinishRegistration error")
	}

	fmt.Printf("%+v \n", user)
	fmt.Printf("%+v \n", sessionData)
	fmt.Printf("%+v \n", c.Request.Body)
	// parsedResponse, err := protocol.ParseCredentialRequestResponseBody(c.Request.Body)
	credential, err := glb.Auth.CreateCredential(user, sessionData.(webauthn.SessionData), response)

	// Handle validation or input errors
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":      "get credential failed: " + err.Error(),
			"verified": false,
		})
		return
	}

	glb.LOG.Info("FinishRegistration success!!!")

	fmt.Printf("FinishRegistration credential: %+v", credential)

	// If login was successful, handle next steps
	// user.AddWebAuthnCredential(*credential)
	// glb.DB.Save(&user)

	glb.LOG.Info("RegisterBegin sessionData.UserID: " + string(user.WebAuthnID()))

	c.JSON(http.StatusOK, gin.H{
		"msg":      "Register Success",
		"verified": true,
	})
}

func LoginBegin(c *gin.Context) {

}
