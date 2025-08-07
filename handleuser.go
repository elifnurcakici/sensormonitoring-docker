package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	// Kullanıcı oturumu kontrolü (örnek)
	session, _ := store.Get(r, "session-name")
	userIDRaw, ok := session.Values["user_id"]
	if !ok {
		http.Error(w, "Yetkisiz erişim", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDRaw.(uint)
	if !ok {
		http.Error(w, "Geçersiz kullanıcı bilgisi", http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Geçersiz istek verisi: "+err.Error(), http.StatusBadRequest)
		return
	}

	var user User
	err = gormDB.First(&user, userID).Error
	if err != nil {
		http.Error(w, "Kullanıcı bulunamadı", http.StatusNotFound)
		return
	}

	// Eski şifre doğrulama
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		http.Error(w, "Eski şifre yanlış", http.StatusUnauthorized)
		return
	}

	// Yeni şifreyi hashle
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Şifre hashlenirken hata oluştu", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedNewPassword)
	err = gormDB.Save(&user).Error
	if err != nil {
		http.Error(w, "Şifre güncellenemedi", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Şifre başarıyla değiştirildi"))
}

type ChangeUsernameRequest struct {
	NewUsername string `json:"new_username"`
}

func handleChangeUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	session, _ := store.Get(r, "session-name")
	userIDRaw, ok := session.Values["user_id"]
	if !ok {
		http.Error(w, "Yetkisiz erişim", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDRaw.(uint)
	if !ok {
		http.Error(w, "Geçersiz kullanıcı bilgisi", http.StatusUnauthorized)
		return
	}

	var req ChangeUsernameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Geçersiz istek verisi: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Kullanıcı adı benzersizliği kontrolü yapılabilir

	var existing User
	err = gormDB.Where("username = ?", req.NewUsername).First(&existing).Error
	if err == nil {
		http.Error(w, "Bu kullanıcı adı zaten kullanılıyor", http.StatusBadRequest)
		return
	}

	var user User
	err = gormDB.First(&user, userID).Error
	if err != nil {
		http.Error(w, "Kullanıcı bulunamadı", http.StatusNotFound)
		return
	}

	user.Username = req.NewUsername
	err = gormDB.Save(&user).Error
	if err != nil {
		http.Error(w, "Kullanıcı adı değiştirilemedi", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Kullanıcı adı başarıyla değiştirildi"))
}
