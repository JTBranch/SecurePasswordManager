#!/bin/bash
#
# Install system dependencies for Fyne applications on Ubuntu/Debian
# This script installs the necessary X11 and OpenGL libraries for building Fyne apps
#

set -e

echo "🔧 Installing system dependencies for Fyne applications..."

# Update package list
echo "📦 Updating package list..."
sudo apt-get update -qq

# Install X11 and OpenGL development libraries
echo "📦 Installing X11 and OpenGL libraries..."
sudo apt-get install -y \
    libgl1-mesa-dev \
    xorg-dev \
    libx11-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libxxf86vm-dev \
    libglu1-mesa-dev

echo "✅ System dependencies installed successfully!"
echo ""
echo "📋 Installed packages:"
echo "  - libgl1-mesa-dev     (OpenGL development files)"
echo "  - xorg-dev            (X.Org development files)"
echo "  - libx11-dev          (X11 client-side library)"
echo "  - libxcursor-dev      (X cursor management library)"
echo "  - libxrandr-dev       (X RandR extension library)"
echo "  - libxinerama-dev     (X Xinerama extension library)"
echo "  - libxi-dev           (X Input extension library)"
echo "  - libxxf86vm-dev      (X Video Mode extension library)"
echo "  - libglu1-mesa-dev    (OpenGL Utility library)"
echo ""
echo "🎯 Ready for Fyne application compilation!"
