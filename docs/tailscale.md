# Tailscale setup

Setting up Tailscale involves installing it on your server and personal devices, connecting them to your private network (called a tailnet), and then accessing services using their private Tailscale names.

## Step 1: Install Tailscale on Each Device

You need to install the Tailscale client on every machine you want to be part of your private network.

- On your Fedora Server:

```bash
# Add the Tailscale repository
sudo dnf config-manager --add-repo https://pkgs.tailscale.com/stable/fedora/tailscale.repo

# Install the package
sudo dnf install tailscale

# Start and enable the service
sudo systemctl enable --now tailscaled
```

On your Laptop, Phone, etc.:

- Go to the official Tailscale downloads page and get the appropriate application for your operating system.

## Step 2: Connect Devices to Your Tailnet

For each device, you need to log in to your Tailscale account to add it to your network.

- On your Fedora Server:

Run the `up` command.

```bash
sudo tailscale up
```

This will give you an authentication URL. Copy this URL and paste it into the browser on your laptop to approve the server.

- On your Laptop or Phone:

Simply open the Tailscale application and click the "Log in" button.

## Step 3: Access Your Services 🚀

Once your devices are connected, they can communicate directly and securely.

1. Find your server's Tailscale name. You can see it in the Tailscale app on your devices or by running sudo tailscale status on the server.

2. Access the service. From your laptop or phone, open a browser and go to `http://<your-vps-name>:<port>`.

## Firewall Troubleshooting (Fedora) 🛡️

If your services are unreachable, the most common issue on Fedora is the firewall blocking the connection. The fix is to tell the firewall to trust all traffic from the secure Tailscale network interface.

```bash
# Add the tailscale0 interface to the 'trusted' zone
sudo firewall-cmd --zone=trusted --add-interface=tailscale0 --permanent

# Reload the firewall to apply the change
sudo firewall-cmd --reload
```

After this, you will not need to open individual ports for any of your private services accessed over Tailscale.
