# SSL/TLS Quick Reference - GoNotes

## 🔒 **Quick SSL Setup (5 menit)**

### 1. **Setup SSL (FREE!)**
```bash
# Dengan Make command
make ssl-setup DOMAIN=example.com EMAIL=admin@example.com

# Dengan script langsung
./scripts/setup_ssl.sh example.com admin@example.com

# Dengan aliases
gn-ssl-setup DOMAIN=example.com EMAIL=admin@example.com
```

### 2. **Start Production dengan SSL**
```bash
# Start semua services dengan SSL
make ssl-start

# Atau dengan aliases
gn-ssl-start
```

### 3. **Check SSL Status**
```bash
# Check certificate info
make ssl-check

# Test SSL online
make ssl-test

# View logs
make ssl-logs
```

## 🚀 **One-Command Production + SSL**

```bash
# Complete production deployment dengan SSL
make prod-setup && \
make ssl-setup DOMAIN=yourdomain.com EMAIL=your-email@example.com && \
make ssl-start
```

## 🔧 **Available SSL Commands**

| Command | Description | Usage |
|---------|-------------|--------|
| `make ssl-setup` | Setup SSL dengan Let's Encrypt | `DOMAIN=example.com EMAIL=admin@example.com` |
| `make ssl-start` | Start production dengan SSL | No parameters |
| `make ssl-stop` | Stop SSL production services | No parameters |
| `make ssl-renew` | Renew SSL certificates | Auto-runs every 12 hours |
| `make ssl-check` | Check certificate status | Optional: `DOMAIN=example.com` |
| `make ssl-test` | Test SSL online (opens URLs) | Optional: `DOMAIN=example.com` |
| `make ssl-logs` | View SSL production logs | No parameters |

## 🌐 **SSL Solutions Comparison**

| Solution | Cost | Setup Time | Auto-Renewal | Wildcard | CDN |
|----------|------|------------|--------------|----------|-----|
| **Let's Encrypt** | FREE | 5 min | ✅ | ✅ | ❌ |
| **Cloudflare** | FREE | 2 min | ✅ | ✅ | ✅ |
| **AWS ACM** | FREE | 3 min | ✅ | ✅ | ❌ |

## 📋 **Prerequisites**

### Domain Requirements
- ✅ Domain name (example.com)
- ✅ Domain pointing to your server IP
- ✅ Port 80 dan 443 open
- ✅ Valid email address

### Server Requirements
- ✅ Docker dan Docker Compose installed
- ✅ Internet connection
- ✅ Root/sudo access (untuk port 80/443)

## 🔍 **SSL Testing URLs**

### Online Testing Tools
```bash
# SSL Labs Test (Grade A-F)
https://www.ssllabs.com/ssltest/analyze.html?d=yourdomain.com

# Security Headers Test
https://securityheaders.com/?q=yourdomain.com

# SSL Certificate Checker
https://www.sslchecker.com/sslchecker/yourdomain.com/443
```

### Command Line Testing
```bash
# Test SSL connection
openssl s_client -connect yourdomain.com:443

# Check certificate expiry
make ssl-check DOMAIN=yourdomain.com

# Test with curl
curl -I https://yourdomain.com
```

## 🚨 **Troubleshooting**

### Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| Domain not pointing | Update DNS A record to server IP |
| Port 80/443 blocked | `sudo ufw allow 80 && sudo ufw allow 443` |
| Certificate expired | `make ssl-renew` |
| Nginx config error | Check logs with `make ssl-logs` |
| Docker not running | `sudo systemctl start docker` |

### Quick Fixes
```bash
# Reset SSL setup
make ssl-stop
rm -rf nginx/ certbot/
make ssl-setup DOMAIN=yourdomain.com EMAIL=your-email@example.com

# Check if ports are open
sudo netstat -tlnp | grep :80
sudo netstat -tlnp | grep :443

# Check DNS resolution
nslookup yourdomain.com
dig yourdomain.com
```

## 💡 **Pro Tips**

### 1. **Automated Renewal**
```bash
# SSL certificates renew automatically every 12 hours
# No manual intervention needed!

# Check renewal logs
make ssl-logs | grep certbot
```

### 2. **Multiple Domains**
```bash
# Support multiple domains in nginx config
server_name yourdomain.com www.yourdomain.com api.yourdomain.com;
```

### 3. **Backup Certificates**
```bash
# Backup SSL certificates
tar -czf ssl-backup-$(date +%Y%m%d).tar.gz certbot/conf/
```

### 4. **Environment Variables**
```bash
# Add to .env.prod for automation
DOMAIN=yourdomain.com
SSL_EMAIL=admin@yourdomain.com
```

## 📊 **Security Score Goals**

### SSL Labs Test Targets:
- ✅ **Grade A** - Excellent SSL configuration
- ✅ **Certificate**: Valid and trusted
- ✅ **Protocol Support**: TLS 1.2 and 1.3 only
- ✅ **Key Exchange**: Strong (ECDHE)
- ✅ **Cipher Strength**: Strong (AES-256-GCM)

### Security Headers Test Targets:
- ✅ **A+** rating
- ✅ HSTS enabled
- ✅ X-Frame-Options: DENY
- ✅ X-Content-Type-Options: nosniff
- ✅ X-XSS-Protection: enabled

## 🎯 **Production Checklist**

### Before SSL Setup
- [ ] Domain DNS pointing to server
- [ ] Server has public IP
- [ ] Ports 80/443 open in firewall
- [ ] Docker running
- [ ] GoNotes app working on HTTP

### After SSL Setup
- [ ] HTTPS loads correctly
- [ ] HTTP redirects to HTTPS
- [ ] SSL Labs test shows Grade A
- [ ] Security headers test shows A+
- [ ] Auto-renewal working
- [ ] Backup certificates created

## 🔗 **Useful Links**

- [Let's Encrypt](https://letsencrypt.org/) - Free SSL certificates
- [SSL Labs Test](https://www.ssllabs.com/ssltest/) - SSL configuration test
- [Security Headers](https://securityheaders.com/) - Security headers test
- [Certbot](https://certbot.eff.org/) - Let's Encrypt client

---

**⏱️ Total setup time: 5-10 minutes**  
**💰 Annual cost: $0 (FREE!)**  
**🔄 Maintenance: Fully automated**  
**🔒 Security: Grade A SSL + A+ Security Headers** 