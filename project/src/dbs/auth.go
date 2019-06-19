// auth
package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/SermoDigital/jose/jws"
)

type Conf struct {
	Method string // 加密算法
	Key    string // 加密key
	Issuer string // 签发者
	Expire int64  // 签名有效期
}

var jwsConf = Conf{
	Method: "HS256",
	Key:    `ddFVDS|{}PDSOIU@$@!!#$$^&^&$`,
	Issuer: "sesan",
	Expire: 5,
}

func getJWS(data map[string]interface{}) (token string, err error) {
	payload := jws.Claims{}
	for k, v := range data {
		payload.Set(k, v)
	}
	now := time.Now()
	payload.SetIssuer(jwsConf.Issuer)
	payload.SetIssuedAt(now)
	payload.SetExpiration(now.Add(time.Duration(jwsConf.Expire) * time.Minute))
	jwtObj := jws.NewJWT(payload, jws.GetSigningMethod(jwsConf.Method))
	tokenBytes, err := jwtObj.Serialize([]byte(jwsConf.Key))
	if err != nil {
		return
	}
	token = string(tokenBytes)
	return
}

func VerifyJWT(token string) (ret bool, err error) {
	jwtObj, err := jws.ParseJWT([]byte(token))
	if err != nil {
		return
	}

	err = jwtObj.Validate([]byte(jwsConf.Key), jws.GetSigningMethod(jwsConf.Method))
	if err == nil {
		ret = true
	}
	return
}

func checkUser(account, password string) bool {
	strSql := "select accout_no, password from account where accout_no=? and password=?"
	user, pwd := "", ""
	if sysdb.QueryRow(strSql, account, password).Scan(&user, &pwd) != nil {
		log.Println(user, pwd)
		return false
	}

	return true
}

func login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	strUser := r.Form.Get("user")
	strPwd := r.Form.Get("pwd")

	w.Header().Set("Server", "nginx")

	if !checkUser(strUser, strPwd) {
		w.WriteHeader(401)
		w.Write([]byte(http.StatusText(401)))
		return
	}

	strJWT, err := getJWS(nil)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Authorization", "Bearer "+strJWT)
}

func mwValidateToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		auths := strings.SplitN(auth, " ", 2)

		w.Header().Set("Server", "nginx")
		if auth == "" || len(auths) != 2 {
			log.Println("auths error:", auths)
			w.Header().Set("WWW-Authenticate", "Bearer")
			w.WriteHeader(401)
			w.Write([]byte(http.StatusText(401)))
			return
		}

		ok, err := VerifyJWT(auths[1])
		if !ok {
			w.WriteHeader(401)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		//w.Header().Set("Access-Control-Expose-Headers", "Authorization,token")
		w.Header().Set("Cache-Control", "no-store")
		next(w, r)
	}
}
