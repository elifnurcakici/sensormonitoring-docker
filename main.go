package main

import (
	"log"
	"net/http"
)

func main() {
	// Veritabanı bağlantısı
	gormDB = connectDB()

	// API endpointleri
	// Manuel veri ekleme endpointleri (sadece login olmuş kullanıcılar için)
	http.Handle("/add/temperature", http.HandlerFunc(handleAddTemperature))
	// http.Handle("/add/temperature", withAuth(http.HandlerFunc(handleAddTemperature))) şeklinde yazarsakta add işleminin yapılabilmesi için login olunmasını ister.
	http.Handle("/add/humidity", http.HandlerFunc(handleAddHumidity))
	http.Handle("/add/pressure", http.HandlerFunc(handleAddPressure))

	http.HandleFunc("/register", handleRegister) //Kullanıcı kayıt
	http.HandleFunc("/login", handleLogin)       // Kullanıcı giriş
	http.HandleFunc("/users", handleGetUsers)    // Kullanıcı listesi
	http.HandleFunc("/ws", handleWebSocket)      // Websocket bağlantısı
	http.HandleFunc("/logout", handleLogout)     // Çıkış

	http.HandleFunc("/data/temperature/all", handleGetTemperature) // sıcaklık
	http.HandleFunc("/data/humidity/all", handleGetHumidity)       // nem
	http.HandleFunc("/data/pressure/all", handleGetPressure)       // basınç

	fs := http.FileServer(http.Dir("./static"))

	// Sadece login.html ve register.html hariç tüm sayfalar korunsun
	http.Handle("/login.html", fs)
	http.Handle("/register.html", fs)

	// Diğer tüm dosyalar için auth zorunlu
	http.Handle("/", withAuth(fs))

	log.Println("Server 8080 portunda çalışıyor...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
