#Install client on sensor

1. scp the client into your home directory on the pi
2. create new service with `sudo nano /lib/systemd/system/eye.service` and add the following:

   ```
   [Unit]
   Description=Eye Service
   After=multi-user.target

   [Service]
   Type=idle
   ExecStart=/usr/bin/python3 /home/pi/client/main.py > /home/pi/eye.log 2>&1

   [Install]
   WantedBy=multi-user.target
   ```

3. give service permission: `sudo chmod 644 /lib/systemd/system/eye.service`
4. reload systemd daemon `sudo systemctl daemon-reload`
5. enable service `sudo systemctl enable sample.service`
6. reboot `sudo reboot`
