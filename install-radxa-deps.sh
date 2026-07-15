#!/usr/bin/env sh
set -eu

sudo apt-get update
sudo apt-get install -y \
  golang \
  build-essential \
  libgl1-mesa-dev \
  libx11-dev \
  libxcursor-dev \
  libxi-dev \
  libxinerama-dev \
  libxrandr-dev \
  libxxf86vm-dev \
  libasound2-dev \
  xserver-xorg \
  xinit \
  openbox
