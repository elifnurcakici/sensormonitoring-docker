package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

// bunu oturum saklamak için kullanıyoruz.
// mesela map[sessionID]role ile sessionID artık anahtar oluyor ve değerde kullanıcı rolü tutuluyor.
var sessions = map[string]string{}

var broadcastChan = make(chan struct{})

func handleRegister(w http.ResponseWriter, r *http.Request) { // Burada sadece POST isteği kabul edilir.
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST isteği desteklenir", http.StatusMethodNotAllowed) // Beklenenden farklı şekilde istek yapılmışsa StatusMethodNotAllowed ile hata döner.
		return
	}

	var req User
	err := json.NewDecoder(r.Body).Decode(&req) // json paketi içindeki NewDecoder ile r.Body alır ve JSON verilerini okumaya hazır hale getirir. Decode ile de okunan veri User structına çevrilir.
	if err != nil {
		http.Error(w, "Geçersiz veri: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Kullanıcı adı ve şifre zorunlu", http.StatusBadRequest)
		return
	}

	// Şifreyi hashle
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost) // bcrypt olarak şifreyi hashler
	if err != nil {
		http.Error(w, "Şifre hashlenemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Password = string(hashedPass) // hashlenmiş şifre

	if req.Role == "" {
		req.Role = "user"
	}

	err = gormDB.Create(&req).Error
	if err != nil {
		log.Println("Kayıt hatası:", err)
		http.Error(w, "Kullanıcı kaydedilemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Yeni kullanıcı kaydı:", req.Username, "role:", req.Role)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Kayıt başarılı"))
}

func handleLogin(w http.ResponseWriter, r *http.Request) { // Burada da sadece POST kabul edilir.
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST isteği desteklenir", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Geçersiz veri: "+err.Error(), http.StatusBadRequest)
		return
	}

	var user User
	err = gormDB.Where("username = ?", req.Username).First(&user).Error // gormDB'nin Where fonksiyonu ile SQL sorgusu oluşturulur.
	// req.Username değeri yerleştirilir ve veritabanında username sütunu req.username değerine eşit olan ilk kayıtı bulnmaya çalışırız.
	if err != nil {
		http.Error(w, "Kullanıcı bulunamadı: "+err.Error(), http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) //user.Password veritabanında kayıtlı olan hashlanmiş veriyi temsil eder
	// req. Pasword ise kullanıcının girişte girdiği o düz text olan şifredir. bcrypteki bu fonksiyon ile req.Password de aynı hashleme algoritması ile hashlenir ve karşılaştırılır
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Kullanıcı adı veya şifre yanlış", http.StatusUnauthorized)
		return
	}

	log.Printf("Giriş başarılı: %s (%s)", user.Username, user.Role)

	sessionID := "sess-" + user.Username // Basit bir sessionID oluşturuyoruz.
	// başına sess ekleyerek
	sessions[sessionID] = user.Role                                                  // en başta tanımladığımız sessions mapinde sessionID ye karşılık olarak admin ya da user olan kullanıcı rolü saklanır
	http.SetCookie(w, &http.Cookie{Name: "session_id", Value: sessionID, Path: "/"}) // session_ID isimli bir cookie oluşturuluyor ve değer olarak session_ID atanıyor.
	// Path: "/" ile bu cookie tüm site için geçerli olur
	// böylece tarayıcı sonraki isteklerde bu cookie'yi sunucuya gönderir ve kullanıcı doğrulanmış olur

	if user.Role == "admin" {
		http.Redirect(w, r, "/indexadmin.html", http.StatusFound) // rol admin ise indexadmin.html sayfasına yönlendirilir.
	} else {
		http.Redirect(w, r, "/indexuser.html", http.StatusFound) // "" user "" indexuser.html """
	}
}

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

func handleGetTemperature(w http.ResponseWriter, r *http.Request) {
	var readings []TemperatureReading // verileri tutacak slice
	if err := gormDB.Order("created_at desc").Limit(10).Find(&readings).Error; err != nil {
		http.Error(w, "Veri alınamadı", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(readings)
}

func handleGetHumidity(w http.ResponseWriter, r *http.Request) {
	var readings []HumidityReading
	if err := gormDB.Order("created_at desc").Limit(10).Find(&readings).Error; err != nil { // en yeni eklenenen göre sıralar son 10 veriyi alır ve reading slicesine atar
		http.Error(w, "Veri alınamadı", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(readings) // readings içindeki veriler JSON' a çevrilir ve HTTPcevabına yazar
}

func handleGetPressure(w http.ResponseWriter, r *http.Request) {
	var readings []PressureReading
	if err := gormDB.Order("created_at desc").Limit(10).Find(&readings).Error; err != nil {
		http.Error(w, "Veri alınamadı", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(readings)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	// Mevcut session ID’yi al
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		delete(sessions, cookie.Value) // Session haritasından da sil
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

// Ortak yardımcı fonksiyon: URL'den id ve value parametrelerini parse eder
func parseIDValue(w http.ResponseWriter, r *http.Request) (uint, float64, bool) {
	idStr := r.URL.Query().Get("id")
	valueStr := r.URL.Query().Get("value")

	if idStr == "" || valueStr == "" {
		http.Error(w, "id ve value parametreleri gerekli", http.StatusBadRequest)
		return 0, 0, false
	}

	// ID'yi uint'e çevir
	idInt, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Geçersiz id", http.StatusBadRequest)
		return 0, 0, false
	}

	// Value'yu float'a çevir
	val, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		http.Error(w, "Geçersiz value", http.StatusBadRequest)
		return 0, 0, false
	}

	return uint(idInt), val, true
}

func handleAddTemperature(w http.ResponseWriter, r *http.Request) {
	id, val, ok := parseIDValue(w, r)
	if !ok {
		return
	}

	record := TemperatureReading{ID: id, Value: val}
	if err := gormDB.Create(&record).Error; err != nil {
		http.Error(w, "Sıcaklık verisi eklenemedi: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Yeni veri eklendi, WebSocket dinleyicilerine bildir
	broadcastChan <- struct{}{}

	w.Write([]byte("Sıcaklık verisi eklendi"))
}

func handleAddHumidity(w http.ResponseWriter, r *http.Request) {
	id, val, ok := parseIDValue(w, r)
	if !ok {
		return
	}

	record := HumidityReading{ID: id, Value: val}
	if err := gormDB.Create(&record).Error; err != nil {
		http.Error(w, "Nem verisi eklenemedi: "+err.Error(), http.StatusBadRequest)
		return
	}
	broadcastChan <- struct{}{}

	w.Write([]byte("Nem verisi eklendi"))
}

func handleAddPressure(w http.ResponseWriter, r *http.Request) {
	id, val, ok := parseIDValue(w, r)
	if !ok {
		return
	}

	record := PressureReading{ID: id, Value: val}
	if err := gormDB.Create(&record).Error; err != nil {
		http.Error(w, "Basınç verisi eklenemedi: "+err.Error(), http.StatusBadRequest)
		return
	}
	broadcastChan <- struct{}{}

	w.Write([]byte("Basınç verisi eklendi"))
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
		if _, ok := sessions[cookie.Value]; !ok {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}

		// Her şey tamamsa devam eder ve sıradaki handler çalıştırılır.
		next.ServeHTTP(w, r)
	})
}
