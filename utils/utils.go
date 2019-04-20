package utils

import (
    "encoding/json"
    "log"
    "net"
    "net/http"
    "net/http/httputil"
    "sync"
    "time"

    "github.com/dgrijalva/jwt-go"
    "github.com/salestock/sersan/config"
)

var (
    num     uint64
    numLock sync.RWMutex
)

type CachedInfo struct {
    Session *SessionInfo
    Proxy   *httputil.ReverseProxy
}

type SessionInfo struct {
    SessionID string
    PodName   string
    Host      string
    Port      string
    BaseURL   string
    VNCPort   string
}

// Serial Serial
func Serial() uint64 {
    numLock.Lock()
    defer numLock.Unlock()
    id := num
    num++
    return id
}

// JsonError JSON error
func JsonError(w http.ResponseWriter, msg string, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(
        map[string]interface{}{
            "value": map[string]string{
                "message": msg,
            },
            "status": 13,
        })
    }

    // SecondsSince Calculate seconds since specified time until now
    func SecondsSince(start time.Time) float64 {
        return float64(time.Now().Sub(start).Seconds())
    }

    // RequestInfo Request info
    func RequestInfo(r *http.Request) (string, string) {
        user := ""
        if u, _, ok := r.BasicAuth(); ok {
            user = u
        } else {
            user = "unknown"
        }
        remote := r.Header.Get("X-Forwarded-For")
        if remote != "" {
            return user, remote
        }
        remote, _, _ = net.SplitHostPort(r.RemoteAddr)
        return user, remote
    }

    // GenerateSessionId Generate Session ID in JWT token format
    func GenerateSessionID(sessionInfo *SessionInfo) (sessionID string, err error) {
        conf := config.Get()
        data := jwt.MapClaims{
            "sessionID": sessionInfo.SessionID,
            "podName":   sessionInfo.PodName,
            "host":      sessionInfo.Host,
            "port":      sessionInfo.Port,
            "baseURL":   sessionInfo.BaseURL,
            "vncPort":   sessionInfo.VNCPort,
        }
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
        sessionID, err = token.SignedString([]byte(conf.SigningKey))
        if err != nil {
            log.Printf("Failed to create formatted session id %v", err)
            return
        }
        log.Printf("Generated Session ID %s", sessionID)
        return
    }

    // ParseSessionID Extract session id information
    func ParseSessionID(sessionID string) (sessionInfo *SessionInfo, err error) {
        conf := config.Get()
        token, err := jwt.Parse(sessionID, func(token *jwt.Token) (interface{}, error) {
            return []byte(conf.SigningKey), nil
        })
        if err != nil {
            log.Printf("Failed to parse session id %v", err)
            return
        }
        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            return &SessionInfo{
                SessionID: claims["sessionID"].(string),
                PodName:   claims["podName"].(string),
                Host:      claims["host"].(string),
                Port:      claims["port"].(string),
                BaseURL:   claims["baseURL"].(string),
                VNCPort:   claims["vncPort"].(string),
            }, nil
        }

        return
    }
