#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "=== HEALTH CHECK ==="
curl -s $BASE_URL/health | jq
echo -e "\n"

# ----------------------------
# STOCK MODULE
# ----------------------------
echo "=== TEST STOCK MODULE ==="

STOCK_NAME="Телефон_$(date +%s)"
echo "Creating stock item: $STOCK_NAME"

curl -s -X POST $BASE_URL/stock/items \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"$STOCK_NAME\",\"stock\":10}" | jq
echo -e "\n"

echo "Listing stock items..."
curl -s $BASE_URL/stock/items | jq
echo -e "\n"

# ----------------------------
# USERS MODULE (опционально)
# ----------------------------
echo "=== TEST USERS MODULE ==="

# Проверяем, доступен ли endpoint /users
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/users)

if [ "$HTTP_CODE" -eq 200 ]; then
    USER_NAME="Иван_$(date +%s)"
    USER_EMAIL="ivan_$(date +%s)@test.com"

    echo "Creating user: $USER_NAME"

    curl -s -X POST $BASE_URL/users \
      -H "Content-Type: application/json" \
      -d "{\"name\":\"$USER_NAME\",\"email\":\"$USER_EMAIL\",\"password\":\"123\",\"role\":\"admin\"}" | jq
    echo -e "\n"

    echo "Listing users..."
    curl -s $BASE_URL/users | jq
    echo -e "\n"
else
    echo "Users module is disabled, skipping..."
fi
