# Makefile for Ngw tool


build: build_ngw build_agent

build_ngw:
	@echo "-> building ngw binary file"
	@go build -o ngw  ngw.go

build_agent:
	@echo "-> building ngw operator file"
	@go build -o ngw_operator  ngw_operator/ngw_operator.go
	@echo "[OK] build binary file successfully"

install:
	@echo "-> agent directory has been created"
	@if [ ! -d /usr/share/ngw ];then mkdir /usr/share/ngw; fi;
	@echo "-> config file directory has been created"
	@if [ ! -d /etc/ngw ];then mkdir /etc/ngw; fi;
	@echo "-> installing agent"
	install -m 755 ngw_operator /usr/share/ngw/ngw_operator
	@echo "-> creating ngw config file"
	install -m 644 ngw.yml /etc/ngw/ngw.yml





