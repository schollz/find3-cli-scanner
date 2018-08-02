FROM ubuntu:18.04
ENV GOLANG_VERSION 1.10
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH /root/go
RUN apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -y libc6-dev make pkg-config g++ gcc git wget wireless-tools bluetooth iw net-tools libpcap-dev && \
	mkdir /root/go && \
	rm -rf /var/lib/apt/lists/* && \
	set -eux; \
	\
# this "case" statement is generated via "update.sh"
	dpkgArch="$(dpkg --print-architecture)"; \
	case "${dpkgArch##*-}" in \
		amd64) goRelArch='linux-amd64'; goRelSha256='b5a64335f1490277b585832d1f6c7f8c6c11206cba5cd3f771dcb87b98ad1a33' ;; \
		armhf) goRelArch='linux-armv6l'; goRelSha256='6ff665a9ab61240cf9f11a07e03e6819e452a618a32ea05bbb2c80182f838f4f' ;; \
		arm64) goRelArch='linux-arm64'; goRelSha256='efb47e5c0e020b180291379ab625c6ec1c2e9e9b289336bc7169e6aa1da43fd8' ;; \
		i386) goRelArch='linux-386'; goRelSha256='2d26a9f41fd80eeb445cc454c2ba6b3d0db2fc732c53d7d0427a9f605bfc55a1' ;; \
		ppc64el) goRelArch='linux-ppc64le'; goRelSha256='a1e22e2fbcb3e551e0bf59d0f8aeb4b3f2df86714f09d2acd260c6597c43beee' ;; \
		s390x) goRelArch='linux-s390x'; goRelSha256='71cde197e50afe17f097f81153edb450f880267699f22453272d184e0f4681d7' ;; \
		*) goRelArch='src'; goRelSha256='f3de49289405fda5fd1483a8fe6bd2fa5469e005fd567df64485c4fa000c7f24'; \
			echo >&2; echo >&2 "warning: current architecture ($dpkgArch) does not have a corresponding Go binary release; will be building from source"; echo >&2 ;; \
	esac; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "${goRelSha256} *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
	if [ "$goRelArch" = 'src' ]; then \
		echo >&2; \
		echo >&2 'error: UNIMPLEMENTED'; \
		echo >&2 'TODO install golang-any from jessie-backports for GOROOT_BOOTSTRAP (and uninstall after build)'; \
		echo >&2; \
		exit 1; \
	fi; \
	\
	export PATH="/usr/local/go/bin:$PATH"; \
	go version && \
	go get -u -v -d github.com/schollz/find3-cli-scanner && \
	go get -u -v -d github.com/google/gopacket/... && \
	cd /root/go/src/github.com/schollz/find3-cli-scanner && \
	git checkout noshark && go build -v && \
	mv find3-cli-scanner /usr/local/bin/ && \
	echo "removing go resources" && rm -rf /usr/local/work/src && \
	echo "purging packages" && apt-get remove -y --auto-remove git libc6-dev pkg-config g++ gcc make && \
	echo "add back pcap" && apt-get update && apt-get install -y libpcap-dev && \
	echo "autoclean" && apt-get autoclean && \
	echo "clean" && apt-get clean && \
	echo "autoremove" && apt-get autoremove -y && \
	echo "rm trash" && rm -rf ~/.local/share/Trash/* && \
	echo "rm go" && rm -rf /usr/local/go* && \
	echo "rm go" && rm -rf /root/go* && \
	echo "rm perl" && rm -rf /usr/share/perl* && \
	echo "rm doc" && rm -rf /usr/share/doc* 


# INSTALL BLUEZ FROM SOURCE
# This is commented out because its not needed except for information's sake
# and maybe if you want to use a smaller image (not Ubuntu:18)
# Instructions: https://github.com/schollz/gatt-python#installing-bluez-from-sources
#RUN  apt-get install -y libusb-dev libdbus-1-dev libglib2.0-dev libudev-dev libical-dev libreadline-dev libdbus-glib-1-dev unzip systemd
#RUN mkdir /root/bluez
#WORKDIR /root/bluez
#RUN wget http://www.kernel.org/pub/linux/bluetooth/bluez-5.9.tar.xz
#RUN tar xf bluez-5.9.tar.xz
#WORKDIR /root/bluez/bluez-5.9
#RUN ./configure --prefix=/usr --sysconfdir=/etc --localstatedir=/var --enable-library
#RUN make
#RUN make install
#RUN ln -svf /usr/libexec/bluetooth/bluetoothd /usr/sbin/
#RUN install -v -dm755 /etc/bluetooth
#RUN install -v -m644 src/main.conf /etc/bluetooth/main.conf
##RUN systemctl daemon-reload
##RUN systemctl start bluetooth


