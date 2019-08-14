#!/bin/bash

if [ -z "$1" ]; then
	echo "Usage: ./CrossCompile.sh <project> [author]"
	exit
fi

author="$2" || "discordconsole-team"

for i in {1..2}; do
	if [ $i == 1 ]
	then arch=386
	else arch=amd64
	fi
	folder="github.com/$author/$1"
	GOOS=linux GOARCH=$arch go install "$folder"
	GOOS=windows GOARCH=$arch go install "$folder"
	GOOS=darwin GOARCH=$arch go install "$folder"
done

dir="$HOME/Downloads/$1"
rm -r "$dir"
mkdir "$dir"

for i in {1..2}; do
	archnum=$((i * 32))
	mkdir -p "$dir/Linux/$archnum-bit"
	mkdir -p "$dir/macOS/$archnum-bit"
	mkdir -p "$dir/Windows/$archnum-bit"
done

for i in {1..2}; do
	archnum=$((i * 32))
	if [ $i == 1 ]; then
		arch=386
		cp "linux_$arch/$1" "$dir/Linux/$archnum-bit/$1"
	else
		arch=amd64
		cp "$1" "$dir/Linux/$archnum-bit/$1"
	fi

	cp "darwin_$arch/$1" "$dir/macOS/$archnum-bit/$1"
	cp "windows_$arch/$1.exe" "$dir/Windows/$archnum-bit/$1.exe"
done

cd "$dir/.."
tar -czf "$1.tar.gz" "$1"
