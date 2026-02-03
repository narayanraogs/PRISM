package server

import (
	"prismServer/global"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// sessionManager safely handles concurrent access to the client map.

type sessionManager struct {
	sync.RWMutex
	clients map[string]*client
}

// Global session manager for all clients.

var sessions = &sessionManager{
	clients: make(map[string]*client),
}

const (
	staleSessionTimeout = 30 * time.Minute
	cleanupInterval     = 5 * time.Minute
)

func init() {
	// Start a background process to clean up stale sessions.
	go sessions.cleanupStaleSessions()
}

func GetRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())

	r.POST("/register", register)
	r.POST("/getValues", getValues)
	r.POST("/setValues", setValues)
	r.GET("/rtUpdate", rtUpdate)
	r.POST("/act", act)
	r.POST("/getTableValues", getTableValues)
	r.GET("/testProgress", testProgressHandler)
	r.POST("/info", remoteInfo)

	return r
}

// getServer is now a thread-safe method.

func (s *sessionManager) getServer(clientID string) *client {
	s.RLock()
	defer s.RUnlock()
	c, ok := s.clients[clientID]
	if ok {
		c.lastSeen = time.Now() // Update lastSeen timestamp on access
		return c
	}
	return nil
}

// createServer is now a thread-safe method.

func (s *sessionManager) createServer(clientID string) *client {
	s.Lock()
	defer s.Unlock()
	c := &client{
		global:   global.ClientGlobal{},
		lastSeen: time.Now(),
	}
	s.clients[clientID] = c
	return c
}

// cleanupStaleSessions iterates through clients and removes inactive ones.

func (s *sessionManager) cleanupStaleSessions() {
	for {
		time.Sleep(cleanupInterval)
		s.Lock()
		for id, c := range s.clients {
			if time.Since(c.lastSeen) > staleSessionTimeout {
				delete(s.clients, id)
			}
		}
		s.Unlock()
	}
}
