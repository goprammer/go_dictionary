#!/bin/bash

# Test REST_API.go
curl -iX GET localhost:8888/v1.0 -H "Content-Type: application/json" -d '{"searchTerm": "dogz", "debug": true}'