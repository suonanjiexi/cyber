package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/suonanjiexi/cyber"
)

// JWTConfig JWT验证中间件配置
type JWTConfig struct {
	SigningKey     string        // JWT签名密钥
	TokenLookup    string        // 从请求中获取token的位置，如 "header:Authorization"
	AuthScheme     string        // 认证方案，如 "Bearer"
	ContextKey     string        // 存储在上下文中的键名
	SigningMethod  string        // 签名方法，如 "HS256"
	TokenHeadName  string        // Token头部名称，如 "Bearer"
	Timeout        time.Duration // Token过期时间
	MaxRefresh     time.Duration // Token最大刷新时间
	TimeFunc       func() time.Time
	IdentityKey    string // 标识键名，如 "id"
	IdentityMethod string // 标识方法，如 "username, email"
}

// DefaultJWTConfig 默认JWT配置
var DefaultJWTConfig = JWTConfig{
	SigningKey:     "cyber_jwt_secret_key",
	TokenLookup:    "header:Authorization",
	AuthScheme:     "Bearer",
	ContextKey:     "user",
	SigningMethod:  "HS256",
	TokenHeadName:  "Bearer",
	Timeout:        time.Hour,
	MaxRefresh:     time.Hour * 24,
	TimeFunc:       time.Now,
	IdentityKey:    "id",
	IdentityMethod: "username",
}

// JWTClaims JWT声明
type JWTClaims struct {
	Id       string                 `json:"id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Role     string                 `json:"role"`
	Exp      int64                  `json:"exp"`
	Iat      int64                  `json:"iat"`
	Custom   map[string]interface{} `json:"custom,omitempty"`
}

// JWTAuth JWT认证中间件
func JWTAuth(next cyber.HandlerFunc) cyber.HandlerFunc {
	return JWTAuthWithConfig(DefaultJWTConfig, next)
}

// JWTAuthWithConfig 使用自定义配置的JWT认证中间件
func JWTAuthWithConfig(config JWTConfig, next cyber.HandlerFunc) cyber.HandlerFunc {
	return func(c *cyber.Context) {
		// 从请求中获取token
		token, err := extractToken(c, config)
		if err != nil {
			c.Error(http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
			return
		}

		// 验证token
		claims, err := validateToken(token, config)
		if err != nil {
			c.Error(http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
			return
		}

		// 将用户信息存储在上下文中
		c.Set(config.ContextKey, claims)

		// 继续处理请求
		next(c)
	}
}

// GenerateToken 生成JWT令牌
func GenerateToken(id, username, email, role string, custom map[string]interface{}, config JWTConfig) (string, error) {
	now := config.TimeFunc().Unix()

	// 创建JWT载荷
	claims := JWTClaims{
		Id:       id,
		Username: username,
		Email:    email,
		Role:     role,
		Iat:      now,
		Exp:      now + int64(config.Timeout.Seconds()),
		Custom:   custom,
	}

	// 编码JWT头部
	header := map[string]interface{}{
		"alg": config.SigningMethod,
		"typ": "JWT",
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerBase64 := base64.RawURLEncoding.EncodeToString(headerBytes)

	// 编码JWT载荷
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// 生成签名
	signatureInput := headerBase64 + "." + payloadBase64
	signature := hmacSha256(signatureInput, config.SigningKey)

	// 组合JWT令牌
	token := headerBase64 + "." + payloadBase64 + "." + signature
	return token, nil
}

// 提取令牌
func extractToken(c *cyber.Context, config JWTConfig) (string, error) {
	parts := strings.Split(config.TokenLookup, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("无效的令牌查找配置: %s", config.TokenLookup)
	}

	extractFrom := parts[0]
	extractKey := parts[1]

	var token string
	switch extractFrom {
	case "header":
		authHeader := c.Request.Header.Get(extractKey)
		if authHeader == "" {
			return "", fmt.Errorf("未提供认证头部")
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != config.TokenHeadName {
			return "", fmt.Errorf("无效的认证头部格式")
		}
		token = parts[1]
	case "query":
		token = c.Request.URL.Query().Get(extractKey)
		if token == "" {
			return "", fmt.Errorf("未提供查询参数令牌")
		}
	case "cookie":
		cookie, err := c.Request.Cookie(extractKey)
		if err != nil || cookie.Value == "" {
			return "", fmt.Errorf("未提供Cookie令牌")
		}
		token = cookie.Value
	default:
		return "", fmt.Errorf("不支持的令牌提取方法: %s", extractFrom)
	}

	return token, nil
}

// 验证令牌
func validateToken(tokenString string, config JWTConfig) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("无效的JWT令牌格式")
	}

	// 验证签名
	signatureInput := parts[0] + "." + parts[1]
	expectedSignature := hmacSha256(signatureInput, config.SigningKey)
	if parts[2] != expectedSignature {
		return nil, fmt.Errorf("无效的JWT签名")
	}

	// 解析载荷
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("无法解码JWT载荷: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("无法解析JWT载荷: %w", err)
	}

	// 验证过期时间
	now := config.TimeFunc().Unix()
	if claims.Exp < now {
		return nil, fmt.Errorf("JWT令牌已过期")
	}

	return &claims, nil
}

// hmacSha256 使用HMAC-SHA256生成签名
func hmacSha256(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
