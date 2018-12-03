package web

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gitgrep-com/gitgrep/config"
)

func jwtCookieAuth(s *Server) *Server {
	cfg := s.cfg
	if cfg.JwtLoginURL == "" || cfg.JwtCookieName == "" || cfg.JwtSecretKey == "" {
		if cfg.JwtLoginURL != "" || cfg.JwtCookieName != "" || cfg.JwtSecretKey != "" {
			log.Printf("Ignoring incomplete configuration for JWT Auth")
		}
		return s
	}
	log.Printf("Using JWT auth with cookie name '%s' and login url: %s",
		cfg.JwtCookieName, cfg.JwtLoginURL)

	sMux := s.mux
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jwtClaims jwt.MapClaims
		cookie, err := r.Cookie(cfg.JwtCookieName)
		if err == nil {
			jwtClaims, err = validateTokenClaims(cfg, cookie.Value)
			if err != nil {
				log.Printf("error in validateToken: %v", err)
			}
		}
		if err != nil || jwtClaims == nil {
			http.Redirect(w, r, cfg.JwtLoginURL, http.StatusSeeOther)
		} else {
			ctx := context.WithValue(r.Context(), "jwt.claims", jwtClaims)
			sMux.ServeHTTP(w, r.WithContext(ctx))
		}
	})
	s.serveWith(h)
	return s
}

func validateTokenClaims(cfg *config.Config, authToken string) (jwt.MapClaims, error) {
	jwToken, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if _, ok := token.Claims.(jwt.MapClaims); ok {
			secretKey, err := base64.StdEncoding.DecodeString(cfg.JwtSecretKey)
			return secretKey, err
		} else {
			return nil, fmt.Errorf("token is missing claims")
		}
	})
	if err != nil {
		log.Printf("Unauthorized: %v", err)
		return nil, fmt.Errorf("unauthorized")
	}
	if claims, ok := jwToken.Claims.(jwt.MapClaims); ok && jwToken.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("unauthorized")
	}
}
