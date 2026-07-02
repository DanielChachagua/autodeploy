#Instalar docker:

~~~
sudo apt-get update
sudo apt-get install ca-certificates curl gnupg lsb-release -y

sudo mkdir -m 0755 -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y

docker --version

sudo groupadd docker
sudo usermod -aG docker $USER
newgrp docker

sudo mkdir -p /etc/systemd/system/docker.service.d/

echo -e "[Service]\nRestart=on-failure\nRestartSec=5s" | sudo tee /etc/systemd/system/docker.service.d/override.conf

sudo systemctl daemon-reload

sudo systemctl restart docker
~~~

# Instalar nginx

~~~
sudo apt update
sudo apt install nginx -y

sudo systemctl status nginx

sudo systemctl enable nginx

sudo mkdir -p /etc/systemd/system/nginx.service.d/

echo -e "[Service]\nRestart=on-failure\nRestartSec=5s" | sudo tee /etc/systemd/system/nginx.service.d/override.conf

sudo systemctl daemon-reload

sudo systemctl start nginx

sudo nano /etc/nginx/sites-available/ejemplo.com

sudo ln -s /etc/nginx/sites-available/ejemplo.com /etc/nginx/sites-enabled/

sudo nginx -t

sudo systemctl reload nginx
~~~

# Instalar cerbot

~~~
sudo apt update
sudo apt install certbot python3-certbot-nginx -y

sudo certbot --nginx -d ejemplo.com -d www.ejemplo.com

sudo certbot renew --dry-run
~~~


# Ejemplo de nginx
~~~
server {
        server_name www.ejemplo.com.ar ejemplo.com.ar;

        client_max_body_size 50M;  # 10MB - ajusta a 50M, 100M, etc.

        location /api/ {
                if ($request_method = 'OPTIONS') {
        add_header 'Access-Control-Allow-Origin' '$http_origin' always;
        add_header 'Access-Control-Allow-Methods' 'GET,POST,PUT,DELETE,OPTIONS,PATCH' always;
        add_header 'Access-Control-Allow-Headers' 'Content-Type,Authorization,Accept,Origin,X-Requested-With,Cache-Control,Pragma' always;
        add_header 'Access-Control-Allow-Credentials' 'true' always;
        add_header 'Access-Control-Max-Age' '86400' always;
        add_header 'Content-Length' '0';
        add_header 'Content-Type' 'text/plain';
        return 204;
                }

                #rewrite ^/api/(.*)$ /$1 break;
                proxy_pass http://localhost:3000;
                proxy_http_version 1.1;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Proto $scheme;
                proxy_read_timeout 60s;
        }

        location /dozzle/ {
                auth_basic "Acceso Restringido - NOA GESTION";
                auth_basic_user_file /etc/nginx/.htpasswd_dozzle;
                #rewrite ^/netdata/(.*)$ /$1 break;
                proxy_pass http://localhost:8888/dozzle/;
                proxy_http_version 1.1;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Proto $scheme;
        }

        location / {
                proxy_pass http://localhost:5000;
                proxy_http_version 1.1;
                proxy_set_header Upgrade $http_upgrade;
                proxy_set_header Connection "upgrade";
                proxy_set_header Host $host;
                proxy_cache_bypass $http_upgrade;
        }

}
~~~

