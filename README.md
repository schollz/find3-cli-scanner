# FIND3 CLI scanner

This is a Bluetooth/WiFi scanner for FIND3. I suggest using Docker for this because it discovers WiFi using `iw` and Docker should make this process platform agnostic (I don't have a Mac so this is the best I can do for you).

# Install

## Quick quick way

Install Docker:

```
$ curl -sSL https://get.docker.com | sh
```

Fetch the latest image:

```
$ docker pull schollz/find3-cli-scanner
```

## Quick way

Install Docker:

```
$ curl -sSL https://get.docker.com | sh
```

And then build the latest image.

```
$ wget https://raw.githubusercontent.com/schollz/find3/master/scanner/Dockerfile
$ docker build -t schollz/find3-cli-scanner .
```

Run the image in the background :

```
$ docker run --net="host" --privileged --name scanning -d -i -t scanner
```

Then, you can send scanning commands using 

**Active scanning**:

```
$ docker exec scanning sh -c "scanner -i YOURINTERFACE -debug -device YOURDEVICE -family YOURFAMILY -server http://YOURSERVER -scantime 10 -bluetooth -forever"
```

**Passive scanning**:

```
$ docker exec scanning sh -c "scanner -i YOURINTERFACE -debug -monitor-mode"
$ docker exec scanning sh -c "scanner -i YOURINTERFACE -debug -device YOURDEVICE -family YOURFAMILY -reverse -server http://YOURSERVER -scantime 10 -bluetooth -forever"
```

See below for more usage.

Start/stop the image using 

```
$ docker start scanning
$ docker stop scanning
```

Jump inside the image:

```
docker run --net="host" --privileged --name scanning -i -t scanner /bin/bash
```


## Usage

### Scan WiFi

```
sudo ./scanner -device YOURCOMPUTER -family YOURFAMILY -i WIFI-INTERFACE 
```

### Scan wifi+bluetooth

```
sudo apt-get install bluez
sudo ./scanner -device YOURCOMPUTER -family YOURFAMILY -i WIFI-INTERFACE -bluetooth
```

### Reverse scan (capture packets of other devices scanning for your computer)

This requires a WiFi card that has promiscuity mode.

If you have two WiFi chips on your computer (one for scanning and one for uploading data) you can do:

```
sudo ./scanner -i wlx98ded0151d38 -set-promiscuous
```

and then

```
sudo ./scanner -device YOURCOMPUTER -family YOURFAMILY -i WIFI-INTERFACE -no-modify -reverse
```

If you only have one WiFi chip on your device, then you can run without `-no-modify`. In this case the WiFi chip will be set/unset after every scan so that it can connect to the internet to upload the packets. This takes about 10 seconds longer though.

```
sudo ./scanner -device YOURCOMPUTER -family YOURFAMILY -i WIFI-INTERFACE -reverse
```
