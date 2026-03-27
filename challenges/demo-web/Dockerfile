FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Install required services
RUN apt-get update && apt-get install -y \
    openssh-server \
    nginx \
    php-fpm \
    php-mysql \
    supervisor \
    curl \
    vim \
    && rm -rf /var/lib/apt/lists/*

# Create Flag file placeholder (will be dynamically overwritten)
RUN touch /flag && chmod 600 /flag

# Create PHP-FPM run directory
RUN mkdir -p /var/run/php

# Configure SSH
RUN mkdir -p /var/run/sshd
RUN echo 'root:awd123456' | chpasswd
RUN sed -i 's/#PermitRootLogin.*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -i 's/#PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config

# Create a simple PHP application with SQL injection vulnerability
RUN mkdir -p /var/www/html
RUN echo '<?php\n\
$host = "localhost";\n\
$user = "root";\n\
$pass = "";\n\
$db = "ctf";\n\
\n\
$conn = new mysqli($host, $user, $pass, $db);\n\
\n\
if (isset(${"_GET"}["id"])) {\n\
    $id = ${"_GET"}["id"];\n\
    // Intentional SQL injection vulnerability\n\
    $sql = "SELECT * FROM users WHERE id = $id";\n\
    $result = $conn->query($sql);\n\
    if ($result) {\n\
        while ($row = $result->fetch_assoc()) {\n\
            echo "ID: " . $row["id"] . "<br>";\n\
            echo "Name: " . $row["name"] . "<br>";\n\
        }\n\
    }\n\
}\n\
?>\n\
<!DOCTYPE html>\n\
<html>\n\
<head><title>AWD Demo Challenge</title></head>\n\
<body>\n\
<h1>Welcome to AWD Demo</h1>\n\
<p>Find the flag in /flag file</p>\n\
<form method="GET">\n\
    <input type="text" name="id" placeholder="Enter ID">\n\
    <button type="submit">Search</button>\n\
</form>\n\
</body>\n\
</html>' > /var/www/html/index.php

# Create supervisor configuration
RUN echo '[supervisord]\n\
nodaemon=true\n\
logfile=/var/log/supervisor/supervisord.log\n\
pidfile=/var/run/supervisord.pid\n\
\n\
[program:sshd]\n\
command=/usr/sbin/sshd -D\n\
autostart=true\n\
autorestart=true\n\
stderr_logfile=/var/log/supervisor/sshd.err.log\n\
stdout_logfile=/var/log/supervisor/sshd.out.log\n\
\n\
[program:nginx]\n\
command=/usr/sbin/nginx -g "daemon off;"\n\
autostart=true\n\
autorestart=true\n\
stderr_logfile=/var/log/supervisor/nginx.err.log\n\
stdout_logfile=/var/log/supervisor/nginx.out.log\n\
\n\
[program:php-fpm]\n\
command=/usr/sbin/php-fpm8.1 -F\n\
autostart=true\n\
autorestart=true\n\
stderr_logfile=/var/log/supervisor/php-fpm.err.log\n\
stdout_logfile=/var/log/supervisor/php-fpm.out.log' > /etc/supervisor/conf.d/supervisord.conf

# Create log directories
RUN mkdir -p /var/log/supervisor

# Configure nginx
RUN echo 'server {\n\
    listen 80;\n\
    root /var/www/html;\n\
    index index.php index.html;\n\
    \n\
    location ~ \.php$ {\n\
        include snippets/fastcgi-php.conf;\n\
        fastcgi_pass unix:/var/run/php/php8.1-fpm.sock;\n\
    }\n\
}' > /etc/nginx/sites-available/default

# Expose ports
EXPOSE 22 80

# Start
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]

