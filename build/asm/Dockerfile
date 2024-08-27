FROM golang:1.22 AS build

WORKDIR /app
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
        git \
 && apt-get clean
RUN go install -v github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest
RUN go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
RUN go install github.com/praetorian-inc/fingerprintx/cmd/fingerprintx@latest
RUN go install github.com/tomnomnom/assetfinder@latest
RUN go install github.com/hahwul/dalfox/v2@latest
RUN git clone --depth 1 https://github.com/robertdavidgraham/masscan.git
RUN git clone --depth 1 https://github.com/projectdiscovery/nuclei-templates.git
RUN cd /app/masscan && make && make install
COPY . .
RUN go mod tidy
RUN go build -o main ./cmd/asm/main.go
FROM ghcr.io/praetorian-inc/noseyparker:v0.19.0 AS noseyparker

FROM debian:stable-slim AS release
WORKDIR /app
RUN apt-get update && apt-get install -y \
    apt-transport-https ca-certificates curl gnupg lsb-release git awscli nmap python3 python3-pip python3-psutil \
    && rm -rf /var/lib/apt/lists/*
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" \
    | tee /etc/apt/sources.list.d/google-cloud-sdk.list
RUN apt-get update && apt-get install -y google-cloud-sdk \
    && apt-get clean && rm -rf /var/lib/apt/lists/*
RUN pip install --break-system-packages git+https://github.com/praetorian-inc/gato.git@main
RUN pip install --break-system-packages azure-cli
RUN az extension add --name resource-graph
RUN touch /tmp/.nuclei-ignore
COPY --from=build /app/masscan/bin/masscan /usr/bin/masscan
COPY --from=build /go/bin/nuclei /usr/bin/nuclei
COPY --from=build /go/bin/subfinder /usr/bin/subfinder
COPY --from=noseyparker /usr/local/bin/noseyparker /usr/bin/noseyparker
COPY --from=build /go/bin/fingerprintx /usr/bin/fingerprintx
COPY --from=build /go/bin/assetfinder /usr/bin/assetfinder
COPY --from=build /go/bin/dalfox /usr/bin/dalfox
COPY --from=build /app/nuclei-templates /root/nuclei-templates
COPY --from=build /app/main /usr/bin/chariot-asm
ENV NUCLEI_TEMPLATES /root/nuclei-templates
ENTRYPOINT ["chariot-asm"]
