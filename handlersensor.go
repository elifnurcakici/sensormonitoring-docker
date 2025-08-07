package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// ---------------------------------- GET -------------------------------------
func handleGetSensorData(w http.ResponseWriter, r *http.Request) {
	sensorIDStr := r.URL.Query().Get("sensor_id")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		http.Error(w, "Geçersiz sensor_id", http.StatusBadRequest)
		return
	}

	var readings []SensorReading
	if err := gormDB.Preload("Sensor").
		Where("sensor_id = ?", sensorID).
		Order("created_at DESC").
		Limit(10).
		Find(&readings).Error; err != nil {
		http.Error(w, "Veri alınamadı", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(readings)
}

func handleGetSensors(w http.ResponseWriter, r *http.Request) {
	sensorType := r.URL.Query().Get("type")
	if sensorType == "" {
		http.Error(w, "type parametresi gerekli", http.StatusBadRequest)
		return
	}
	session, _ := store.Get(r, "session-name")
	workspaceID, ok := session.Values["workspace_id"].(uint)
	if !ok {
		http.Error(w, "Workspace ID alınamadı", http.StatusUnauthorized)
		return
	}

	var sensors []Sensor
	if err := gormDB.Where("type = ? AND workspace_id = ?", sensorType, workspaceID).Find(&sensors).Error; err != nil {
		http.Error(w, "Sensörler alınamadı", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sensors)
}

// ---------------------------------- ADD -------------------------------------
func handleAddSensorReading(w http.ResponseWriter, r *http.Request) {
	sensorIDStr := r.URL.Query().Get("sensor_id")
	valueStr := r.URL.Query().Get("value")

	if sensorIDStr == "" || valueStr == "" {
		http.Error(w, "Eksik parametre: sensor_id ve value gereklidir", http.StatusBadRequest)
		return
	}

	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		http.Error(w, "Geçersiz sensor_id", http.StatusBadRequest)
		return
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		http.Error(w, "Geçersiz value", http.StatusBadRequest)
		return
	}

	var sensor Sensor
	if err := gormDB.First(&sensor, sensorID).Error; err != nil {
		http.Error(w, "Sensör bulunamadı", http.StatusNotFound)
		return
	}

	reading := SensorReading{
		SensorID:  sensor.ID,
		Value:     value,
		CreatedAt: time.Now(),
	}

	if err := gormDB.Create(&reading).Error; err != nil {
		http.Error(w, "Veri eklenemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Veri başarıyla eklendi!"))
}

// ---------------------------------- UPDATE -------------------------------------

func handleUpdateSensorReading(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	sensorIDStr := r.URL.Query().Get("sensor_id")
	valueStr := r.URL.Query().Get("value")

	if idStr == "" || sensorIDStr == "" || valueStr == "" {
		http.Error(w, "Eksik parametre: id, sensor_id ve value gereklidir", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Geçersiz id", http.StatusBadRequest)
		return
	}

	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		http.Error(w, "Geçersiz sensor_id", http.StatusBadRequest)
		return
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		http.Error(w, "Geçersiz value", http.StatusBadRequest)
		return
	}

	var reading SensorReading
	if err := gormDB.Where("id = ? AND sensor_id = ?", id, sensorID).First(&reading).Error; err != nil {
		http.Error(w, "Veri bulunamadı", http.StatusNotFound)
		return
	}

	reading.Value = value
	if err := gormDB.Save(&reading).Error; err != nil {
		http.Error(w, "Veri güncellenemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Veri başarıyla güncellendi"))
}

// ---------------------------------- DELETE -------------------------------------

func handleDeleteSensorReading(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	sensorIDStr := r.URL.Query().Get("sensor_id")

	if idStr == "" || sensorIDStr == "" {
		http.Error(w, "Eksik parametre: id ve sensor_id gereklidir", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Geçersiz id", http.StatusBadRequest)
		return
	}

	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		http.Error(w, "Geçersiz sensor_id", http.StatusBadRequest)
		return
	}

	if err := gormDB.Where("id = ? AND sensor_id = ?", id, sensorID).Delete(&SensorReading{}).Error; err != nil {
		http.Error(w, "Veri silinemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Veri başarıyla silindi"))
}
func handleGetAccessibleSensors(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	userIDRaw := session.Values["user_id"]
	workspaceIDRaw := session.Values["workspace_id"]

	userID, ok1 := userIDRaw.(uint)
	_, ok2 := workspaceIDRaw.(uint)

	if !ok1 || !ok2 {
		http.Error(w, "Kullanıcı oturumu bulunamadı", http.StatusUnauthorized)
		return
	}

	var user User
	if err := gormDB.First(&user, userID).Error; err != nil {
		http.Error(w, "Kullanıcı bilgisi alınamadı: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var sensors []Sensor

	if user.Role == "admin" {
		// Admin tüm sensörleri görebilir
		err := gormDB.Find(&sensors).Error
		if err != nil {
			http.Error(w, "Sensörler alınamadı: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Sadece kendi workspace'ine ait sensörler
		err := gormDB.Where("workspace_id = ?", user.WorkspaceID).Find(&sensors).Error
		if err != nil {
			http.Error(w, "Sensörler alınamadı: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sensors)
}
