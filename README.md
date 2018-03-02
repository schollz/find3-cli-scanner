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

## Natively

I don't recommed this because I can't gaurantee that all the processes that the scanner calls will work in every OS. I can tell you that these instructions will work on Ubuntu16/18 though.

Install the dependencies.

```
$ sudo apt-get install wireless-tools iw net-tools
```

(Optional) If you want to do Bluetooth scanning too, then also:

```
$ sudo apt-get install bluetooth
```

(Optional) If you want to do Passive scanning, then do:

```
$ sudo apt-get install tshark
```

Now [Install Go](https://golang.org/dl/) and pull the latest:

```
$ go get -u -v github.com/schollz/find3-cli-scanner
```

# Usage 

## Docker usage

First start the docker container in the background.

```
$ docker run --net="host" --privileged --name scanner -d -i -t schollz/find3-cli-scanner
```

Then, you can send scanning commands using 

**Active scanning**:

```
$ docker exec scanner sh -c "find3-cli-scanner -i YOURINTERFACE -debug -device YOURDEVICE -family YOURFAMILY -server http://YOURSERVER -scantime 10 -bluetooth -forever"
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
