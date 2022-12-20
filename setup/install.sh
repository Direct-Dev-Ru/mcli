#!/usr/bin/env bash
# This is the entry point for configuring the system.
#####################################################

#install basic tools
sudo DEBIAN_FRONTEND=noninteractive apt-get -y update
sudo DEBIAN_FRONTEND=noninteractive apt-get -y upgrade
#sudo DEBIAN_FRONTEND=noninteractive apt-get -y -o Dpkg::Options::="--force-confdef" \
#-o Dpkg::Options::="--force-confnew" install git

#get golang
wget https://go.dev/dl/go1.19.linux-amd64.tar.gz

#unzip the archive
tar -xvf go1.19.linux-amd64.tar.gz

#move the go lib to local folder
mv go /usr/local

#delete the source file
rm go1.19.linux-amd64.tar.gz

#only full path will work
touch /home/vagrant/.bash_profile

echo "export PATH=$PATH:/usr/local/go/bin" >>/home/vagrant/.bash_profile

echo "export GOPATH=/home/vagrant/workspace:$PATH" >>/home/vagrant/.bash_profile

export GOPATH=/home/vagrant/workspace

mkdir -p "$GOPATH/bin"
