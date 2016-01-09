#!/bin/bash
# Ubuntu 15.10 setup shell script

sudo apt-get update
sudo apt-get upgrade

# install essential packages
sudo apt-get -y install build-essential git nano vim-nox wget curl htop dstat debian-archive-keyring software-properties-common

# nginx 1.9
sudo add-apt-repository ppa:nginx/development

# golang 1.5
sudo add-apt-repository ppa:evarlast/golang1.5

# install
sudo apt-get update
sudo apt-get install -y nginx golang
