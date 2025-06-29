# Bender Reverse Proxy 🌀

Golang tabanlı, dinamik konfigürasyon destekli bir reverse proxy uygulamasıdır.  
Yük dengeleme, temel kimlik doğrulama, path rewrite, health check ve hot-reload özelliklerini destekler.

## 🚀 Özellikler

- 🔁 **Load Balancing (Round Robin)**
- 🔐 **Basic Auth Desteği**
- ✏️ **Path Rewrite Mekanizması**
- ❤️ **Health Check Kontrolü**
- 🔄 **Hot Reload (YAML dosyası izlenir)**
- 🧪 **Test Edilebilirlik (unit test, integration test)**
- 🐳 **Docker ile kolay dağıtım**
- ⚙️ **CI/CD (GitHub Actions üzerinden)**

## 🗂️ Proje Yapısı

bender-reverse-proxy/
├── api-backend/               # Örnek hedef API (Merhaba + Health endpoint içerir)
│   ├── backend.go             # Basit Go ile yazılmış HTTP sunucu
│   └── Dockerfile             # API için Docker imajı oluşturur
├── router.go                  # Reverse proxy yönlendirme mantığı (auth, rewrite, rr, health vs.)
├── main.go                    # Uygulamanın ana çalıştırma dosyası
├── routes.yaml                # Proxy route'larının YAML formatında tanımı (dinamik hot reload destekli)
├── router_test.go             # Unit test dosyaları (auth, rewrite, round robin, health)
└── .github/
    └── workflows/
        └── ci.yml             # GitHub Actions CI pipeline (test ve docker build içerir)




## 🧰 Kurulum ve Çalıştırma


# Projeyi klonla
git clone https://github.com/yusufbender/bender-reverse-proxy.git
cd bender-reverse-proxy

# API backend'i docker ile ayağa kaldır
cd api-backend
docker build -t my-api .
docker run -d -p 5001:5678 my-api
docker run -d -p 5003:5678 my-api

# Reverse proxy başlat
cd ..
go run main.go

## routes.yaml Örneği
routes:
  - path: /api/user
    targets:
      - http://localhost:5001
      - http://localhost:5003
    rewrite: /user
    auth:
      username: admin
      password: 1234

## Test Çalıştırma

go test -v

## ⚙️ CI/CD
Push edildiğinde testler ve Docker build işlemleri otomatik çalışır:
.github/workflows/ci.yml dosyasını içerir.

🧠 Geliştirici: Yusuf Bender
Yazılım geliştirici, IT & DevOps meraklısı.
Daha fazlası: [LinkedIn](https://www.linkedin.com/in/yusufbender/)


## 📌 Notlar
Daha fazla özellik için issue oluşturabilirsiniz.

Projeye katkı sağlamak istersen PR açmaktan çekinme.
