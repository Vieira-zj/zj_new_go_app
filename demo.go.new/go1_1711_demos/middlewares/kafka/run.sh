#!/bin/bash
set -eu

function product_messages {
    echo "product kafka messages by access http server."
    for i in $(seq 30 40); do
        curl "http://localhost:8080/?data${i}"
        # curl "http://127.0.0.1:8080/?data${i}"
    done
}

if [[ $1 == "product" ]]; then
    product_messages
    exit 0
fi

echo "done"
