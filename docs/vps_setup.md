# VPS Setup and Hardening Guide

This guide covers the end-to-end process of setting up a secure Fedora-based VPS, from initial login to deploying applications behind an Nginx reverse proxy with SSL. It includes common troubleshooting steps encountered along the way.

---

## 1. Initial Server Access & User Setup

Always avoid using the default cloud provider user (`fedora`, `ubuntu`, `ec2-user`, etc.) for daily operations. Creating a personal, non-root user with `sudo` privileges is a critical security best practice.

### 1.1 Create a New User

Log in as the default `fedora` user and create your new personal user.

```bash
# Replace 'your_user' with your desired username
sudo adduser your_user
```

### 1.2 Grant Sudo Privileges

On Fedora, administrative rights are granted by adding the user to the `wheel` group.

```bash
sudo usermod -aG wheel your_user
```

### 1.3 Set a Password for the New User

This is necessary for the new user to use `sudo`.

```bash
sudo passwd your_user
```

## 2. SSH Key Authentication Setup

Password logins are insecure and vulnerable to brute-force attacks. We will set up SSH key authentication for our new user and then disable passwords entirely.

### 2.1 Generate SSH Keys (On Your Local Computer)

If you don't already have an SSH key, create one on your personal computer. The Ed25519 algorithm is modern and recommended.

```bash
# Press Enter to accept defaults for file location and no passphrase
ssh-keygen -t ed25519 -C "your_email@example.com"
```

### 2.2 Copy the Public Key to the Server

Because most secure cloud images disable password authentication by default, the `ssh-copy-id` command often fails. The best method is to copy the key manually while you are already logged in as the `fedora` user.

Run these commands on the VPS:

```bash
# 1. Create the .ssh directory for your new user
sudo mkdir /home/your_user/.ssh

# 2. Copy the authorized_keys file from the 'fedora' user
#    This file already contains the public key you used to log in
sudo cp /home/fedora/.ssh/authorized_keys /home/your_user/.ssh/authorized_keys

# 3. CRITICAL: Fix ownership of the new directory and file
sudo chown -R your_user:your_user /home/your_user/.ssh

# 4. CRITICAL: Set strict permissions required by SSH
sudo chmod 700 /home/your_user/.ssh
sudo chmod 600 /home/your_user/.ssh/authorized_keys
```

You can now log out and log back in directly as your new user with your existing SSH key: `ssh your_user@YOUR_VPS_IP`.

## 3. Harden the SSH Server 🛡️

Now that key-based login is confirmed for your new user, disable password authentication for the entire server.

### 3.1 Edit the SSHD Config File

```bash
sudo vim /etc/ssh/sshd_config
```

Find the line `PasswordAuthentication` yes and change it to `no`. If it's commented out with a `#`, remove the `#`.

```toml
PasswordAuthentication no
```

### 3.2 Restart the SSH Service

Apply the changes by restarting the service.

```bash
sudo systemctl restart sshd.service
```

## SSH Troubleshooting

* **Error:** `Unit sshd.service could not be found`
    * **Cause:** Your distribution uses `ssh.service` instead of `sshd.service`. This is the standard on Debian-based systems like Ubuntu.
    * **Fix:** Use the correct service name: `sudo systemctl restart ssh.service`.

* **Problem:** Can't log in with a password even after setting `PasswordAuthentication yes` in the config file.
    * **Cause:** The SSH server only reads its configuration file when it starts. Any changes you make will not take effect until the service is restarted.
    * **Fix:** Restart the SSH service to apply the new settings: `sudo systemctl restart sshd.service`.

## 4. Nginx Reverse Proxy Setup

To host multiple apps, use Nginx as a reverse proxy to direct traffic based on the domain name.

### 4.1 Install and Enable Nginx

```bash
sudo dnf install nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

### 4.2 Configure the Firewall

On Fedora, you must explicitly allow web traffic through `firewalld`.

```bash
sudo firewall-cmd --add-service=http --permanent
sudo firewall-cmd --add-service=https --permanent
sudo firewall-cmd --reload
```

### 4.3 Create an Nginx Server Block

Create a new configuration file for your site in `/etc/nginx/conf.d/`.

```bash
sudo vim /etc/nginx/conf.d/your_domain.conf
```

Paste this configuration to route traffic to an app running on `localhost:8000`.

```bash
server {
    listen 80;
    listen [::]:80;

    server_name your_domain.com www.your_domain.com;

    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 4.4 Configure SELinux

This is a required step on Fedora to allow Nginx to make network connections to your backend application.

```bash
sudo setsebool -P httpd_can_network_connect 1
```

### 4.5 Test and Reload Nginx

```bash
# Test for syntax errors
sudo nginx -t

# If successful, reload the service to apply the new config
sudo systemctl reload nginx
```

## 5. Securing Your Site with SSL (Let's Encrypt) 🔒

Use Certbot to automatically get and renew a free SSL certificate.

### 5.1 Install Certbot

```bash
sudo dnf install certbot python3-certbot-nginx
```

### 5.2 Run Certbot

This command will automatically detect your domains from the Nginx config, get a certificate, and update the config file for you.

```bash
sudo certbot --nginx
```

### Troubleshooting Certbot
* **Error:** `no valid A records found` or `NXDOMAIN looking up A`
    * **Cause:** This is a DNS error. It means your domain name is not pointing to your server's IP address. The Certificate Authority cannot find your server to verify ownership.
    * **Fix:** Log in to your **domain registrar** (e.g., GoDaddy, Namecheap) and create two **"A" records**:
        * **Host:** `@`, **Value:** `YOUR_VPS_IP_ADDRESS`
        * **Host:** `www`, **Value:** `YOUR_VPS_IP_ADDRESS`
    * Wait for DNS to propagate (this can take a few minutes to an hour), then run `sudo certbot --nginx` again.

## 6. Managing Environment Variables

To configure your applications without hard-coding secrets, use environment variables.

1. Open your user's bash configuration file:

```bash
vim ~/.bashrc
```

2. Add `export` lines to the bottom of the file:

```bash
export NOSTRICH_WATCH_DB_PASSWORD="your_db_password_string"
export NOSTRICH_WATCH_MONITOR_PRIVATE_KEY="your_secret_key"
```

3. Load the new variables into your current session so you don't have to log out and back in:

```bash
source ~/.bashrc
```

Any application started by this user will now have access to these variables.
