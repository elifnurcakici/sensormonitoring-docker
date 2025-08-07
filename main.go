package main

import (
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("gizli-bir-anahtar"))

type RecaptchaResponse struct {
	Success bool `json:"success"`
}

func main() {
	// Veritabanı bağlantısı
	gormDB = connectDB()

	// API endpointleri
	// Manuel veri ekleme endpointleri (sadece login olmuş kullanıcılar için)
	// http.Handle("/add/temperature", withAuth(http.HandlerFunc(handleAddTemperature))) şeklinde yazarsakta add işleminin yapılabilmesi için login olunmasını ister.
	http.HandleFunc("/add", handleAddSensorReading)
	http.HandleFunc("/update", handleUpdateSensorReading)
	http.HandleFunc("/delete", handleDeleteSensorReading)

	http.HandleFunc("/register", handleRegister) //Kullanıcı kayıt
	http.HandleFunc("/login", handleLogin)       // Kullanıcı giriş
	http.HandleFunc("/users", handleGetUsers)    // Kullanıcı listesi
	http.HandleFunc("/ws", handleWebSocket)      // Websocket bağlantısı
	http.HandleFunc("/logout", handleLogout)     // Çıkış
	// r.Handle("/api/user/change-credentials", authMiddleware(handleChangeCredentials(sensordb))).Methods("POST")
	http.HandleFunc("/api/sensors", handleGetSensors)
	http.HandleFunc("/api/sensor-data", handleGetSensorData)
	http.HandleFunc("/api/user-info", handleGetUserInfo)
	http.HandleFunc("/api/accessible-sensors", handleGetAccessibleSensors)

	http.HandleFunc("/api/user-workspaces", handleGetUserWorkspaces)
	http.HandleFunc("/api/sensor-counts", handleGetSensorCounts)

	http.HandleFunc("/change-password", handleChangePassword)
	http.HandleFunc("/change-username", handleChangeUsername)

	fs := http.FileServer(http.Dir("./static"))

	// Sadece login.html ve register.html hariç tüm sayfalar korunsun
	http.Handle("/login.html", fs)
	http.Handle("/register.html", fs)
	http.HandleFunc("/sensor", serveSensorPage)

	http.HandleFunc("/api/login-history", handleLoginHistory)

	// Diğer tüm dosyalar için auth zorunlu
	http.Handle("/", withAuth(fs))

	log.Println("Server 8080 portunda çalışıyor...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

/*func verifyRecaptcha(token string) bool {
	secret := "6LfKkZUrAAAAAMMP-5DgeqFUg3nG87Yo3rZHvPUr" // Google’dan aldığın Secret Key

	data := url.Values{}
	data.Set("secret", secret)
	data.Set("response", token)

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", data)
	if err != nil {
		log.Println("reCAPTCHA isteği başarısız:", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result RecaptchaResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println("reCAPTCHA JSON çözümleme hatası:", err)
		return false
	}

	return result.Success
}
*/
