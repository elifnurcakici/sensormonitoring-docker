package main

import (
	"encoding/json"
	"net/http"
)

// bunu oturum saklamak için kullanıyoruz.
// mesela map[sessionID]role ile sessionID artık anahtar oluyor ve değerde kullanıcı rolü tutuluyor.
var sessionRoles = map[string]string{}

var broadcastChan = make(chan struct{})

func handleGetUsers(w http.ResponseWriter, r *http.Request) { // Burada ise sadece GET istekleri kabul edilir.
	if r.Method != http.MethodGet {
		http.Error(w, "Sadece GET isteği desteklenir", http.StatusMethodNotAllowed)
		return
	}

	var users []User
	if err := gormDB.Find(&users).Error; err != nil { // Find(&users) ile User tablosundaki tüm kullanıcı kayıtları users slicesinde tutulur
		http.Error(w, "Kullanıcılar alınamadı: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parola bilgisini döndürmemek için şifre alanı temizlenir.
	// bu sayede sunucuda da görünmez
	for i := range users {
		users[i].Password = ""
	}

	w.Header().Set("Content-Type", "application/json") // HTTP yanıtının içerik tipini JSON olarak belirleriz.
	json.NewEncoder(w).Encode(users)                   // users slicesindeki tüm kullanıcılar JSON formatına çevrilir ve HTTP cevabına yazdırılır.

	// ---- API endpointleri ----

}

func withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Giriş sayfaları kontrol edilmez
		if r.URL.Path == "/login.html" || r.URL.Path == "/register.html" {
			next.ServeHTTP(w, r)
			return
		}

		//tarayıcıdan gelen cookie alınır.
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login.html", http.StatusFound) // Eğer yoksa veya değeri boşsa kullanıcı login.html sayfasına yönlendirilir
			return
		}

		// Cookie değeri sessions map’te kayıtlı değilse giriş yaptırma
		if _, ok := sessionRoles[cookie.Value]; !ok {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}

		// Her şey tamamsa devam eder ve sıradaki handler çalıştırılır.
		next.ServeHTTP(w, r)
	})
}
