# 📝 GoNotes - Simple Setup

## 🔧 Development

```bash
# 1. Clone project
git clone <your-repo>
cd gonotes

# 2. Start database & redis
docker-compose -f docker-compose.dev.yaml up -d

# 3. Install dependencies 
go mod tidy

# 4. Run migrations
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/gonotes_dev?sslmode=disable" up

# 5. Start app (hot reload)
air
```

**Test:** `curl http://localhost:8081/health`

---

## 🚀 Production

```bash
# 1. Clone project
git clone <your-repo>
cd gonotes

# 2. Setup environment
cp .env.example .env
nano .env  # Edit dengan nilai production

# 3. Setup domain di nginx.conf
sed -i 's/YOUR_DOMAIN/yourdomain.com/g' nginx.conf

# 4. Run production
docker-compose -f docker-compose.prod.yaml up -d
```

**Setup SSL:**
```bash
# Generate SSL certificate (setelah docker running)
docker-compose -f docker-compose.prod.yaml run --rm certbot certonly --webroot --webroot-path /var/www/certbot/ -d yourdomain.com --email admin@yourdomain.com --agree-tos --no-eff-email
```

**Test:** `curl https://yourdomain.com/health`

---

## ⚙️ Important Files

- **Dockerfile** - Universal container (dev/prod)
- **docker-compose.dev.yaml** - Database + Redis only 
- **docker-compose.prod.yaml** - Full production stack
- **nginx.conf** - Reverse proxy + SSL
- **.env.example** - Environment template

**DNS Setup:** Point `yourdomain.com` A record to your VPS IP

That's it! 🎉