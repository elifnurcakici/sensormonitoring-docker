Gerçek Zamanlı Sensör İzleme Uygulaması

Bu proje, Go dilinde geliştirilmiş, PostgreSQL veritabanı ve WebSocket teknolojisi kullanarak gerçek zamanlı sıcaklık, nem ve basınç verilerini izleyebileceğiniz bir web uygulamasıdır.

Özellikler
- Gerçek zamanlı sensör verilerinin anlık görüntülenmesi
- Kullanıcı girişi ve yetkilendirme
- WebSocket üzerinden anlık veri iletimi
- Docker ile kolay kurulum ve çalıştırma

Teknolojiler
- Go (Golang)
- PostgreSQL
- WebSocket
- Docker & Docker Compose
- HTML, CSS, JavaScript


Kurulum ve Çalıştırma
1. Projeyi klonlayın:
git clone https://github.com/elifnurcakici/sensormonitoring-docker.git
cd sensormonitoring-docker

2. Docker ile projeyi çalıştırın:
docker-compose up --build -d

3. Tarayıcınızda http://localhost:8080 adresine gidin.

## 🌐 Login Sayfası
![Login](static/screenshoots/login.jpeg)

## 🌐 Register Sayfası
![Register](static/screenshoots/register.jpeg)

## 🌐 User Ana Sayfası
![User](static/screenshoots/indexadmin.jpeg)

## 🌐 Admin Ana Sayfası
![Admin](static/screenshoots/login.jpeg)

## 🌐 Veri ekleme
![Add](static/screenshoots/add.jpeg)

## 🌐 Sıcaklık Sayfası
![Temperature](static/screenshoots/temperature.png)

## 🌐 Nem Sayfası
![Humidity](static/screenshoots/humidity.png)

## 🌐 Basınç Sayfası
![Pressure](static/screenshoots/pressure.png)

API Endpointleri
- GET /data/temperature/all - Son 10 sıcaklık verisi
- GET /data/humidity/all - Son 10 nem verisi
- GET /data/pressure/all - Son 10 basınç verisi
- GET /ws - WebSocket bağlantısı


