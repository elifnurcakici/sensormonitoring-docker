package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var req struct {
	Username string `json:"username"`
	Password string `json:"password"`
	// Token    string `json:"token"`
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST isteği desteklenir", http.StatusMethodNotAllowed)
		return
	}

	type RegisterRequest struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		Role        string `json:"role"`
		WorkspaceID uint   `json:"workspace_id"`
		Token       string `json:"token"`
	}

	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Geçersiz veri: "+err.Error(), http.StatusBadRequest)
		return
	}

	/*if !verifyRecaptcha(req.Token) {
		http.Error(w, "reCAPTCHA doğrulaması başarısız", http.StatusUnauthorized)
		return
	}
	*/

	// Zorunlu alanları kontrol et
	if req.Username == "" || req.Password == "" || req.WorkspaceID == 0 {
		http.Error(w, "Kullanıcı adı, şifre ve workspace seçimi zorunludur", http.StatusBadRequest)
		return
	}

	// Şifre hashle
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Şifre hashlenemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user := User{
		Username:    req.Username,
		Password:    string(hashedPass),
		Role:        req.Role,
		WorkspaceID: req.WorkspaceID,
	}

	if user.Role == "" {
		user.Role = "user"
	}

	err = gormDB.Create(&user).Error
	if err != nil {
		log.Println("Kayıt hatası:", err)
		http.Error(w, "Kullanıcı kaydedilemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Yeni kullanıcı kaydı: %s, role: %s, workspaceID: %d\n", user.Username, user.Role, user.WorkspaceID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Kayıt başarılı"))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST isteği desteklenir", http.StatusMethodNotAllowed)
		return
	}

	// Gelen istekteki JSON verisini karşılayacak struct

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Geçersiz veri: "+err.Error(), http.StatusBadRequest)
		return
	}

	// reCAPTCHA doğrulaması
	/*valid := verifyRecaptcha(req.Token)
	if !valid {
		http.Error(w, "reCAPTCHA doğrulaması başarısız", http.StatusUnauthorized)
		return
	}*/

	// Kullanıcıyı veritabanından bul
	var user User
	err = gormDB.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		http.Error(w, "Kullanıcı bulunamadı", http.StatusUnauthorized)
		return
	}

	// Şifre karşılaştırması
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		http.Error(w, "Kullanıcı adı veya şifre yanlış", http.StatusUnauthorized)
		return
	}

	// Session başlat veya al
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Session hatası: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Session içine kullanıcı bilgilerini ekle
	session.Values["user_id"] = user.ID
	session.Values["workspace_id"] = user.WorkspaceID
	session.Values["role"] = user.Role

	// Giriş kaydı oluştur
	loginLog := LoginLog{
		UserID:  user.ID,
		LoginAt: time.Now(),
	}
	if err := gormDB.Create(&loginLog).Error; err == nil {
		session.Values["login_log_id"] = loginLog.ID
	}

	// Session'u kaydet (bu, otomatik olarak cookie'yi set eder)
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session kaydedilemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Giriş başarılı: %s (%s)", user.Username, user.Role)

	// JSON olarak redirect bilgisini döndür
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"redirect": "/dashboard.html",
	})
}

func handleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	userID, ok := session.Values["user_id"].(uint)
	if !ok {
		http.Error(w, "Kullanıcı oturumu bulunamadı", http.StatusUnauthorized)
		return
	}

	var user User
	if err := gormDB.First(&user, userID).Error; err != nil {
		http.Error(w, "Kullanıcı bulunamadı", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           user.ID,
		"username":     user.Username,
		"workspace_id": user.WorkspaceID,
		"role":         user.Role,
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	sessionS, _ := store.Get(r, "session-name")

	// login_log_id varsa logout zamanını güncelle
	if logIDRaw, ok := sessionS.Values["login_log_id"]; ok {
		if logID, ok := logIDRaw.(uint); ok {
			var loginLog LoginLog
			if err := gormDB.First(&loginLog, logID).Error; err == nil {
				now := time.Now()
				loginLog.LogoutAt = &now
				gormDB.Save(&loginLog)
			}
		}
	}

	// Session'ı temizle
	sessionS.Options.MaxAge = -1
	sessionS.Save(r, w)

	// Mevcut session ID’yi al
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		delete(sessionRoles, cookie.Value) // Session haritasından da sil
	}

	// Cookie’yi tarayıcıdan sil
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	log.Println("Kullanıcı çıkış yaptı")

	// Login sayfasına yönlendir
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

func handleLoginHistory(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	userIDRaw := session.Values["user_id"]
	userID, ok := userIDRaw.(uint)
	if !ok || userID == 0 {
		http.Error(w, "Giriş yapılmamış", http.StatusUnauthorized)
		return
	}

	var logs []LoginLog
	err := gormDB.Where("user_id = ?", userID).
		Order("login_at DESC").
		Limit(5).
		Find(&logs).Error

	if err != nil {
		http.Error(w, "Kayıtlar alınamadı: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type Entry struct {
		Time   string `json:"time"`
		Action string `json:"action"`
	}

	var entries []Entry
	for _, log := range logs {
		entries = append(entries, Entry{
			Time:   log.LoginAt.Local().Format("2006-01-02 15:04:05"),
			Action: "Giriş",
		})
		if log.LogoutAt != nil {
			entries = append(entries, Entry{
				Time:   log.LogoutAt.Local().Format("2006-01-02 15:04:05"),
				Action: "Çıkış",
			})
		}
	}

	// Zaman sırasına göre tekrar sırala (yeniye göre)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time > entries[j].Time
	})

	if len(entries) > 5 {
		entries = entries[:5]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func handleGetUserWorkspaces(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := getUserFromSession(r)
	if !ok {
		http.Error(w, "Yetkisiz", http.StatusUnauthorized)
		return
	}

	var workspaces []Workspace

	if role == "admin" {
		// Admin tüm workspace'leri görebilir
		if err := gormDB.Find(&workspaces).Error; err != nil {
			http.Error(w, "Veritabanı hatası", http.StatusInternalServerError)
			return
		}
	} else {
		// Normal kullanıcı sadece kendi workspace'ini görebilir
		var user User
		if err := gormDB.First(&user, userID).Error; err != nil {
			http.Error(w, "Kullanıcı bulunamadı", http.StatusNotFound)
			return
		}

		var workspace Workspace
		if err := gormDB.First(&workspace, user.WorkspaceID).Error; err != nil {
			http.Error(w, "Workspace bulunamadı", http.StatusNotFound)
			return
		}

		workspaces = append(workspaces, workspace)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workspaces": workspaces,
	})
}

func handleGetSensorCounts(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := getUserFromSession(r)
	if !ok {
		http.Error(w, "Yetkisiz", http.StatusUnauthorized)
		return
	}

	// Kullanıcının workspace_id'sini al
	var user User
	if err := gormDB.First(&user, userID).Error; err != nil {
		http.Error(w, "Kullanıcı bulunamadı", http.StatusNotFound)
		return
	}

	// O workspace'e bağlı sensörleri say
	var count int64
	if err := gormDB.Model(&Sensor{}).Where("workspace_id = ?", user.WorkspaceID).Count(&count).Error; err != nil {
		http.Error(w, "Veritabanı hatası", http.StatusInternalServerError)
		return
	}

	response := map[string]int64{
		"sensorCount": count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getUserFromSession(r *http.Request) (uint, string, bool) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		return 0, "", false
	}

	userID, ok := session.Values["user_id"].(uint)
	if !ok {
		return 0, "", false
	}

	role, ok := session.Values["role"].(string)
	if !ok {
		return 0, "", false
	}

	return userID, role, true
}

func serveSensorPage(w http.ResponseWriter, r *http.Request) {
	user, role, ok := getUserFromSession(r)
	if !ok || role != "admin" {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	log.Printf("Admin kullanıcı %s sensör sayfasına erişiyor", user)
	http.ServeFile(w, r, "./static/sensor.html")
}
