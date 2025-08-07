package main

/*
- Websocket bağlantısı açılıyor
- Her 2 saniyede veritabanından en güncel sıcaklık,nem ve basınç değerlerş çekiliyor.
- Bu veriler JSON formatında Websocket ile anlık olarak istemciye gönderiliyor.
- Bağlantı kapanır ya da hata oluşursa da duruyor. */

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{ // Wensocket bağlantısına HTTP isteğinden geçiş için
	CheckOrigin: func(r *http.Request) bool { return true }, // CheckOrigin ile origin başlığını kontrol ederiz.
	// Burası şimdilik bütün domainlerden gelen websocket bağlantı isteğini kabul ediyor.
	// ancak biz bunu sadece belirli domainlerden gelen isteklere çevirebiliriz.
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade hatası:", err)
		return
	}
	defer ws.Close()

	for {
		select {
		case <-broadcastChan:
			// Tüm sensörleri al
			var sensors []Sensor
			if err := gormDB.Find(&sensors).Error; err != nil {
				log.Println("Sensörler alınamadı:", err)
				continue
			}

			data := make(map[string]interface{})

			// Her bir sensör için son veriyi çek
			for _, sensor := range sensors {
				var reading SensorReading
				err := gormDB.Where("sensor_id = ?", sensor.ID).
					Order("created_at desc").
					First(&reading).Error
				if err != nil {
					log.Println("Okuma alınamadı:", err)
					continue
				}
				data[sensor.Name] = reading
			}

			// Verileri JSON olarak gönder
			if err := ws.WriteJSON(data); err != nil {
				log.Println("WebSocket yazma hatası:", err)
				return
			}
		}
	}
}
