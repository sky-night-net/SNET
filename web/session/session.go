package session

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const (
	SessionName = "SNET_SESSION"
	UserKey     = "session_user"
)

func NewStore() sessions.Store {
	// In a real app, the secret should be in settings/env
	store := cookie.NewStore([]byte("snet-secret-key-12345"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24 * 7, // 7 days
		HttpOnly: true,
	})
	return store
}

func Get(c *gin.Context, key interface{}) interface{} {
	session := sessions.Default(c)
	return session.Get(key)
}

func Set(c *gin.Context, key interface{}, val interface{}) error {
	session := sessions.Default(c)
	session.Set(key, val)
	return session.Save()
}

func Clear(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	return session.Save()
}

func IsLogin(c *gin.Context) bool {
	return Get(c, UserKey) != nil
}
