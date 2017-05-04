// authentication
package main

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"time"

	"datastruct/safemap"
)

var (
	//seconds
	Check_Interval time.Duration = 10
	//seconds
	KnockOutTime int64 = 60 * 60
)

var signMap *safemap.SafeMap

func init() {
	signMap = safemap.NewSafeMap()
	go func() {
		timer1 := time.NewTicker(Check_Interval * time.Second)
		for {
			select {
			case <-timer1.C:
				knockout()
			}
		}
	}()
}

func loop(k, v interface{}) bool {
	if time.Now().Unix()-v.(int64) >= KnockOutTime {
		return true
	}
	return false
}

func knockout() {
	signMap.LoopDel(loop)
}

func genToken(guestIP, user string) (token string, ok bool) {
	strSign := guestIP + user
	if strings.TrimSpace(strSign) == "" {
		return "", false
	}

	byte_md5 := md5.Sum([]byte(strSign))
	return hex.EncodeToString(byte_md5[:]), true
}

//AddAuth login.
func AddAuth(guestIP, user, pass string) bool {
	strSign, ok := genToken(guestIP, user)
	if !ok {
		return false
	}

	if signMap.Check(strSign) {
		return true
	}

	if !SCUCheck(user, pass) {
		return false
	}

	signMap.Set(strSign, time.Now().Unix())
	return true
}

//CheckAuth
func CheckAuth(guestIP, user string) bool {
	strSign, ok := genToken(guestIP, user)
	if !ok {
		return false
	}
	//return signMap.Check(strSign)
	return signMap.CheckWithUpdate(strSign, time.Now().Unix())
}

//RemoveAuth logout.
func RemoveAuth(guestIP, user string) {
	strSign, ok := genToken(guestIP, user)
	if !ok {
		return
	}

	signMap.Del(strSign)
}
