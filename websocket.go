package main

/*
- Websocket bağlantısı açılıyor
- Her 2 saniyede veritabanından en güncel sıcaklık,nem ve basınç değerlerş çekiliyor.
- Bu veriler JSON formatında Websocket ile anlık olarak istemciye gönderiliyor.
- Bağlantı kapanır ya da hata oluşursa da duruyor. */

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{ // Wensocket bağlantısına HTTP isteğinden geçiş için
	CheckOrigin: func(r *http.Request) bool { return true }, // CheckOrigin ile origin başlığını kontrol ederiz.
	// Burası şimdilik bütün domainlerden gelen websocket bağlantı isteğini kabul ediyor.
	// ancak biz bunu sadece belirli domainlerden gelen isteklere çevirebiliriz.
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) { //HTTP isteği geldiğinde bu fonksiyon Websocket bağlantısını kurup veri gönderir.
	ws, err := upgrader.Upgrade(w, r, nil) // HTTP bağlantısı WebSocket bağlantısına dönüştürülür.
	if err != nil {
		log.Println("WebSocket upgrade hatası:", err) // Başarısız olursa burada loglanıyor.
		return
	}
	defer ws.Close()

	ticker := time.NewTicker(5 * time.Second) // burada kanal her 5 saniyede bir tetikleniyor.
	// Her 5 saniyede bir veri göndermek için bu teteikleyiciyi kullanırız.
	defer ticker.Stop()

	for { // sonsuz for döngüsü başlar
		select {
		case <-broadcastChan:
			var temp TemperatureReading
			var hum HumidityReading
			var press PressureReading

			// Son kayıtları çek
			gormDB.Order("created_at desc").First(&temp)  // burada TemperatureReading tablsoundan en son oluşturulan (created at' i en büyük olan)kaydı alıyor.
			gormDB.Order("created_at desc").First(&hum)   // ""   HumidityReading tablsonndan
			gormDB.Order("created_at desc").First(&press) // ""  PressureReading

			// JSON olarak gönderilmek için bir map oluşturulur.
			data := map[string]interface{}{
				"temperature": temp,
				"humidity":    hum,
				"pressure":    press,
			}

			// Göndermek için
			err := ws.WriteJSON(data) //WriteJSON ile datayı JSOn formatına çevirip WebSocket üzerinden gönderir
			if err != nil {
				break //Eğer hata olursa döngü kapanır.
			}
		}
	}
}
