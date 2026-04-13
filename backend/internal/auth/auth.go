package auth

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret []byte
	hrPin  string
}

func NewManager(secret, hrPin string) *Manager {
	return &Manager{secret: []byte(secret), hrPin: hrPin}
}

// Claims is the JWT payload carried by authenticated requests.
type Claims struct {
	UserID string   `json:"uid"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	HRGate bool     `json:"hr_gate,omitempty"`
	jwt.RegisteredClaims
}

func (m *Manager) Issue(userID, email string, roles []string, ttl time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(m.secret)
}

func (m *Manager) Parse(raw string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

const ctxClaimsKey = "auth.claims"

// Middleware validates the bearer token and stashes claims on the gin context.
func (m *Manager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		claims, err := m.Parse(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(ctxClaimsKey, claims)
		c.Next()
	}
}

func ClaimsFrom(c *gin.Context) (*Claims, bool) {
	v, ok := c.Get(ctxClaimsKey)
	if !ok {
		return nil, false
	}
	cl, ok := v.(*Claims)
	return cl, ok
}

// RequireRole aborts if the caller does not have any of the allowed roles.
func RequireRole(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFrom(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		for _, want := range allowed {
			for _, have := range claims.Roles {
				if have == want {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

// RequireHRGate ensures the token has passed the HR PIN check.
func RequireHRGate() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFrom(c)
		if !ok || !claims.HRGate {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "HR PIN required"})
			return
		}
		c.Next()
	}
}

// VerifyPin handler: accepts { "pin": "..." } and issues a short-lived
// HR-gated token derived from the current claims.
func (m *Manager) VerifyPin(c *gin.Context) {
	var body struct {
		PIN string `json:"pin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if subtle.ConstantTimeCompare([]byte(body.PIN), []byte(m.hrPin)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bad pin"})
		return
	}
	claims, ok := ClaimsFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	newClaims := &Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Roles:  claims.Roles,
		HRGate: true,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	signed, err := tok.SignedString(m.secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sign"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": signed, "expires_in": 1800})
}
