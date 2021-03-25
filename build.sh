#!/usr/bin/env bash

set -x

GitCommitLog=`git log --pretty=oneline -n 1`
GitCommitLog=${GitCommitLog//\'/\"}
GitStatus=`git status -s`

LDFlags=" \
    -X 'github.com/aiot-network/aiot-network/version.GitCommitLog=${GitCommitLog}' \
    -X 'github.com/aiot-network/aiot-network/version.GitStatus=${GitStatus}' \
"

ROOT_DIR=`pwd`
CHAIN_DIR=`pwd`"/cmd/chain"
WALLET_DIR=`pwd`"/cmd/wallet"
BOOT_DIR=`pwd`"/cmd/boot"
rm -rf bin
mkdir bin

cd ${CHAIN_DIR} && GOOS=linux GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/linux/aiotchain/chain &&
cd ${CHAIN_DIR} && GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/darwin/aiotchain/chain &&
cd ${CHAIN_DIR} && GOOS=windows GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/windows/aiotchain/chain.exe &&
cp ${CHAIN_DIR}/config.toml ${ROOT_DIR}/bin/linux/aiotchain/ &&
cp ${CHAIN_DIR}/config.toml ${ROOT_DIR}/bin/darwin/aiotchain/ &&
cp ${CHAIN_DIR}/config.toml ${ROOT_DIR}/bin/windows/aiotchain/ &&

cd ${WALLET_DIR} && GOOS=linux GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/linux/wallet/wallet &&
cd ${WALLET_DIR} && GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/darwin/wallet/wallet &&
cd ${WALLET_DIR} && GOOS=windows GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/windows/wallet/wallet.exe &&
cp ${WALLET_DIR}/wallet.toml ${ROOT_DIR}/bin/linux/wallet/ &&
cp ${WALLET_DIR}/wallet.toml ${ROOT_DIR}/bin/darwin/wallet/ &&
cp ${WALLET_DIR}/wallet.toml ${ROOT_DIR}/bin/windows/wallet/ &&

cd ${BOOT_DIR} && GOOS=linux GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/linux/boot/boot &&
cd ${BOOT_DIR} && GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/darwin/boot/boot &&
cd ${BOOT_DIR} && GOOS=windows GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/windows/boot/boot.exe &&

Version=`${ROOT_DIR}/bin/darwin/aiotchain/chain --version`

cd ${ROOT_DIR} &&
zip -r bin/${Version}-linux-amd64.zip ./bin/linux &&
zip -r bin/${Version}-darwin-amd64.zip ./bin/darwin &&
zip -r bin/${Version}-windows-amd64.zip ./bin/windows &&


ls -lrt ${ROOT_DIR}/bin/linux/aiotchain &&
ls -lrt ${ROOT_DIR}/bin/linux/wallet &&
ls -lrt ${ROOT_DIR}/bin/linux/boot &&

ls -lrt ${ROOT_DIR}/bin/darwin/aiotchain &&
ls -lrt ${ROOT_DIR}/bin/darwin/wallet &&
ls -lrt ${ROOT_DIR}/bin/darwin/boot &&

ls -lrt ${ROOT_DIR}/bin/windows/aiotchain &&
ls -lrt ${ROOT_DIR}/bin/windows/wallet &&
ls -lrt ${ROOT_DIR}/bin/windows/boot &&
echo 'build done.'