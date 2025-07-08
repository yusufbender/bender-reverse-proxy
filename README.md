# Bender Reverse Proxy

Basit, Go tabanlı bir reverse proxy uygulaması.  
Dinamik konfigürasyon ile yük dengeleme, kimlik doğrulama, path rewrite, health check ve hot-reload gibi temel ihtiyaçları doğrudan çözer.

## Temel Özellikler

- Round-robin yük dengeleme
- Basic Auth (kullanıcı adı/şifre koruması)
- Path rewrite
- Health check (servis izleme)
- Hot reload (config dosyası izleniyor)
- Birim ve entegrasyon testleri
- Docker ile dağıtım
- CI/CD (GitHub Actions ile otomasyon)

## Kurulum ve Kullanım

```bash
git clone https://github.com/yusufbender/bender-reverse-proxy.git
cd bender-reverse-proxy

# API backend örneğini başlatmak için:
cd api-backend
docker build -t my-api .
docker run -d -p 5001:5678 my-api
docker run -d -p 5003:5678 my-api
cd ..

# Reverse proxy başlat
go run main.go
# Test
go test -v
