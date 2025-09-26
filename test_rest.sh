#!/bin/bash

BASE_URL="http://0.0.0.0:8080/api/v1"

echo "=== HEALTH CHECK ==="
curl -s $BASE_URL/health | jq
echo

# --- Categories ---
echo "=== TEST CATEGORIES ==="
curl -s -X POST $BASE_URL/stock/categories -H "Content-Type: application/json" -d '{"name":"Электроника"}' | jq
curl -s $BASE_URL/stock/categories | jq
echo

# --- Units ---
echo "=== TEST UNITS ==="
curl -s -X POST $BASE_URL/stock/units -H "Content-Type: application/json" -d '{"name":"шт"}' | jq
curl -s $BASE_URL/stock/units | jq
echo

# --- Warehouses ---
echo "=== TEST WAREHOUSES ==="
curl -s -X POST $BASE_URL/stock/warehouses -H "Content-Type: application/json" -d '{"name":"Основной склад","address":"Москва"}' | jq
curl -s $BASE_URL/stock/warehouses | jq
echo

# --- Counterparties ---
echo "=== TEST COUNTERPARTIES ==="
curl -s -X POST $BASE_URL/stock/counterparties -H "Content-Type: application/json" -d '{"name":"Поставщик 1"}' | jq
curl -s $BASE_URL/stock/counterparties | jq
echo

# --- Items ---
echo "=== TEST ITEMS ==="
curl -s -X POST $BASE_URL/stock/items -H "Content-Type: application/json" -d '{"name":"Телефон","sku":"TEL001","unit_id":1,"category_id":1,"price":10000}' | jq
curl -s $BASE_URL/stock/items | jq
echo

# --- Stock Movements ---
echo "=== TEST STOCK MOVEMENTS ==="
curl -s -X POST $BASE_URL/stock/movements -H "Content-Type: application/json" -d '{"item_id":1,"warehouse_id":1,"counterparty_id":1,"quantity":10,"type":"in","comment":"Поставка"}' | jq
curl -s $BASE_URL/stock/movements | jq
echo

# --- Documents ---
echo "=== TEST DOCUMENTS ==="
curl -s -X POST $BASE_URL/stock/documents -H "Content-Type: application/json" -d '{"type":"Invoice","number":"INV001","warehouse_id":1,"counterparty_id":1,"comment":"Документ поставки","items":[{"item_id":1,"quantity":10}]}' | jq
curl -s $BASE_URL/stock/documents | jq
echo
