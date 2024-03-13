#!/bin/bash -e

TMP=$(mktemp --directory)

# Windows
pushd $TMP
wget https://s3.amazonaws.com/session-manager-downloads/plugin/latest/windows/SessionManagerPlugin.zip
unzip SessionManagerPlugin.zip
unzip package.zip

popd
cp $TMP/bin/session-manager-plugin.exe assets/plugin/windows_amd64
rm -rf $TMP/*

# Mac AMD64
pushd $TMP
wget https://s3.amazonaws.com/session-manager-downloads/plugin/latest/mac/sessionmanager-bundle.zip
unzip sessionmanager-bundle.zip

popd
cp $TMP/sessionmanager-bundle/bin/session-manager-plugin assets/plugin/darwin_amd64
rm -rf $TMP/*

# Mac ARM
pushd $TMP
wget https://s3.amazonaws.com/session-manager-downloads/plugin/latest/mac_arm64/sessionmanager-bundle.zip
unzip sessionmanager-bundle.zip

popd
cp $TMP/sessionmanager-bundle/bin/session-manager-plugin assets/plugin/darwin_arm64
rm -rf $TMP/*

# Linux AMD64
pushd $TMP
wget https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb
dpkg-deb -x session-manager-plugin.deb .

popd
cp $TMP/usr/local/sessionmanagerplugin/bin/session-manager-plugin assets/plugin/linux_amd64
rm -rf $TMP/*

# Linux ARM64
pushd $TMP
wget https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_arm64/session-manager-plugin.deb
dpkg-deb -x session-manager-plugin.deb .

popd
cp $TMP/usr/local/sessionmanagerplugin/bin/session-manager-plugin assets/plugin/linux_arm64
rm -rf $TMP/*
