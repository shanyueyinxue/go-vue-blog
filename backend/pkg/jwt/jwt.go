package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID     uint   `json:"userId"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	PwdVersion uint   `json:"pwdVersion"`
	jwt.RegisteredClaims
}

type JwtConfig struct {
	Secret            string `mapstructure:"secret"`
	Issuer            string `mapstructure:"issuer"`
	Expiration        int    `mapstructure:"expiration"`
	RefreshExpiration int    `mapstructure:"refreshExpiration"`
}
type JWT struct {
	Config JwtConfig
}

func NewJWT(config JwtConfig) *JWT {
	return &JWT{Config: config}
}

var (
	ErrInvalidToken     = errors.New("无效的token")
	ErrSignatureInvalid = errors.New("签名无效")
	ErrInvalidClaims    = errors.New("无效的claims")
)

const REFRESH_ISSUER_PREFIX = "refresh:"

// 常见弱密钥黑名单（生产环境直接拒绝启动）
var weakSecrets = map[string]bool{
	"":           true,
	"secret":     true,
	"change-me":  true,
	"change_me":  true,
	"password":   true,
	"123456":     true,
	"1234567890": true,
	"jwt-secret": true,
}

// ValidateSecret 校验 JWT 密钥强度：拒绝弱密钥与过短密钥（< 16 字节）。
func ValidateSecret(secret string) error {
	if weakSecrets[secret] || len(secret) < 16 {
		return errors.New("JWT 密钥过短或使用了默认弱密钥，请配置至少 16 位的随机密钥（jwt.secret）")
	}
	return nil
}

func (j *JWT) GetRefreshIssuer() string {
	return REFRESH_ISSUER_PREFIX + j.Config.Issuer
}

func (j *JWT) GenerateToken(uid uint, username, email string, pwdVersion uint) (string, error) {
	secret := j.Config.Secret
	expiration := j.Config.Expiration
	// refreshExpiration :=  j.Config.RefreshExpiration
	issuer := j.Config.Issuer
	now := time.Now()
	claims := JWTClaims{
		UserID:     uid,
		Username:   username,
		Email:      email,
		PwdVersion: pwdVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiration) * time.Second)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (j *JWT) GenerateRefreshToken(uid uint, username, email string, pwdVersion uint) (string, error) {
	secret := j.Config.Secret
	expiration := j.Config.RefreshExpiration
	issuer := j.GetRefreshIssuer()
	now := time.Now()
	claims := JWTClaims{
		UserID:     uid,
		Username:   username,
		Email:      email,
		PwdVersion: pwdVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiration) * time.Second)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (j *JWT) ParseToken(tokenString string) (*JWTClaims, error) {
	secret := j.Config.Secret
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !token.Valid {
		return nil, ErrSignatureInvalid
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}
	return claims, nil
}

// 获取token的过期时间
func (j *JWT) GetTokenExpiration(tokenString string) (*time.Time, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}
	return &claims.ExpiresAt.Time, nil
}
