#!/bin/bash

TEMPD=""

add_systemd() {
    curl -L -o /etc/systemd/system/ddns-go.service https://raw.githubusercontent.com/domaingts/ddns-go/refs/heads/master/script/ddns-go.service
}

add_config() {
    curl -L -o /etc/ddns-go/config.yaml https://raw.githubusercontent.com/domaingts/ddns-go/refs/heads/master/script/config.yaml
}

make_configuration_folder() {
  mkdir -p /etc/ddns-go
}

download() {
    TEMPD="$(mktemp -d)"
    local temp_file
    temp_file="$(mktemp)"
    if ! curl -sS -H "Accept: application/vnd.github.v3+json" -o "$temp_file" 'https://api.github.com/repos/ddns-go/realm/releases/latest'; then
        "rm" "$temp_file"
        echo 'error: Failed to get release list, please check your network.'
    fi
    version="$(sed 'y/,/\n/' "$temp_file" | grep 'tag_name' | awk -F '"' '{print $4}')"
    "rm" "$temp_file"
    local package="ddns-go-linux-amd64-v3.tar.gz"
    echo 'https://github.com/domaingts/ddns-go/releases/download/$version/$package'
    curl -L --retry 5 --retry-delay 10 --retry-max-time 60 "https://github.com/domaingts/ddns-go/releases/download/$version/$package" -o "$TEMPD/$package"
    tar Cxzvf "$TEMPD" "$TEMPD/$package"
    location="$TEMPD/ddns-go"
    mv "$location" /usr/local/bin/
    ddns-go -v | tee
}

main() {
    add_systemd
    make_configuration_folder
    add_config
    download
}

main