FROM golang:1.21.6-bullseye

SHELL ["/bin/bash", "-c"]

RUN sed -i 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list
RUN sed -i 's/security.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list

RUN apt-get update && apt-get -y upgrade \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get -y clean

WORKDIR /build

COPY . .

ENV GOPROXY="https://goproxy.cn"

RUN go mod tidy \
    && go mod download \
    && go build -o /extproc


FROM busybox

COPY --from=0 /extproc /bin/extproc
RUN chmod +x /bin/extproc

ARG EXAMPLE=allow-and-block

EXPOSE 50051

ENTRYPOINT [ "/bin/extproc" ]

# CMD [ "allow-and-block", "--log-stream", "--log-phases", "--blocklist", "192.168.1.2" ]
CMD [ "allow-and-block", "--log-stream", "--log-phases", "--allowlist", "192.168.1.2" ]
