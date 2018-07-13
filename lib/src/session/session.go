// session project session.go
package session

import (
	"log"
	"snowflake"
	_ "snowflake"
	"strconv"
	"sync"
	"time"
)

var sfw *snowflake.Worker

type Session struct {
	sid          string    //唯一标示
	lastAccessed int64     //最后访问时间
	maxAge       int64     //超时时间
	data         *sync.Map //主数据
}

func init() {
	sw, err := snowflake.NewWorker(1)
	if err != nil {
		log.Panic(err)
	}

	sfw = sw
}

func newSession() *Session {
	return &Session{
		sid:          strconv.FormatInt(sfw.NextID(), 10),
		lastAccessed: time.Now().Unix(),
		maxAge:       60 * 30, //默认30分钟
		data:         &sync.Map{},
	}
}

func (si *Session) Set(key, value interface{}) {
	if !si.valid() {
		return
	}
	si.data.Store(key, value)
}

func (si *Session) Get(key interface{}) interface{} {
	if !si.valid() {
		return nil
	}
	val, rst := si.data.Load(key)
	if rst {
		return val
	}
	return nil
}

func (si *Session) Del(key interface{}) {
	si.data.Delete(key)
}

func (si *Session) SID() string {
	return si.sid
}

func (si *Session) active() {
	si.lastAccessed = time.Now().Unix()
}

func (si *Session) valid() bool {
	if time.Now().Unix()-si.lastAccessed < si.maxAge {
		return true
	} else {
		return false
	}
}

type SessionMgr struct {
	sessionMap *sync.Map //主数据
}

//******************************************
func NewSessionMgr() *SessionMgr {
	return &SessionMgr{
		sessionMap: &sync.Map{},
	}
}

func (sm *SessionMgr) Get(k interface{}) *Session {
	ss, rst := sm.sessionMap.Load(k)
	if rst && ss.(*Session).valid() {
		si := ss.(*Session)
		si.active()
		return si
	}

	return nil
}

func (sm *SessionMgr) NewSessino() *Session {
	si := newSession()
	sm.sessionMap.Store(si.sid, si)
	return si
}

func (sm *SessionMgr) Del(k interface{}) {
	sm.sessionMap.Delete(k)
}
