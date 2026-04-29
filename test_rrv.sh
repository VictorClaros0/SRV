#!/bin/bash

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"usuario":"admin","contrasena":"admin123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

echo "=== TOKEN ==="
echo $TOKEN

echo ""
echo "=== PRUEBA 1: Subir imagen (RRV → MongoDB) ==="
curl -s -X POST http://localhost:8080/api/rrv/actas/upload \
  -F "file=@/home/anett-garcia/Downloads/SRV/SRV/backend/actas/acta_1010200001001.pdf" \
  -F 'acta_data={"acta_id":"ACTA-1010200001001","departamento":"Chuquisaca","municipio":"Yotala","recinto":"U.E. Padresama","mesa":"1010200001001","candidatos":[{"candidato_id":"CAND-01","nombre":"Daenerys Targaryen","votos":140},{"candidato_id":"CAND-02","nombre":"Sansa Stark","votos":39},{"candidato_id":"CAND-03","nombre":"Robert Baratheon","votos":124},{"candidato_id":"CAND-04","nombre":"Tyrion Lannister","votos":345}],"votos_nulos":64,"votos_blancos":76,"total_votos":788}'


echo ""
echo "=== PRUEBA 3: Ver actas RRV guardadas en MongoDB ==="
curl -s http://localhost:8080/api/rrv/actas

echo ""
echo "=== PRUEBA 4: Cargar CSV oficial (PostgreSQL) ==="
echo "acta_id,departamento,provincia,municipio,recinto,mesa,candidato_1,candidato_2,candidato_3,candidato_4,votos_nulos,votos_blancos,total_votos,ciudadanos_habilitados
ACTA-001,Cochabamba,Tiraque,Tiraque,Recinto 1,10101001003,100,80,60,40,20,15,315,400" > /tmp/prueba.csv

curl -s -X POST http://localhost:8080/api/oficial/csv/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/tmp/prueba.csv"

echo ""
echo "=== PRUEBA 5: Ver actas oficiales para Pablo ==="
curl -s http://localhost:8080/api/oficial/actas \
  -H "Authorization: Bearer $TOKEN"
