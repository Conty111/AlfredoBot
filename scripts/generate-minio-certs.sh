#!/bin/bash
set -e

# удаляем существующие сертификаты
rm -f *.pem

# генерируем certificate authority с ключом 
# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=XX/ST=XXX/L=XXXX/O=XXXXX/OU=XXXXXX/CN=*.local/emailAddress=mail@gmail.com"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# генерируем сертификат с ключом
# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -keyout server-key.pem -out server-req.pem -subj "/C=XX/ST=XXX/L=XXXX/O=XXXXX/OU=XXXXXX/CN=*.local/emailAddress=mail@gmail.com" -nodes

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in server-req.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -days 60 -extfile certificates-extra.cnf

echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text

echo "Command to verify Server and CA certificates"
openssl verify -CAfile ca-cert.pem server-cert.pem

# удаляем директорию certs, в которую мы положим вновь созданные сертификаты для minio
echo "Deleting certs dir"
rm -rf ../certs

# создаем заново директорию ../certs и помещаем туда необходимые сертификаты
echo "Creating certs dirs"
mkdir -p ../certs/CAs
cp ca-cert.pem ../certs/ca.crt
cp ca-cert.pem ../certs/CAs/ca.crt
cp server-cert.pem ../certs/public.crt
cp server-key.pem ../certs/private.key 
chmod +r ../certs/private.key

rm -f *.pem
rm -f *.srl