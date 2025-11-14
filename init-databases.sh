#!/bin/bash

# **What it does:**
# - Creates all databases when PostgreSQL starts
# - Grants permissions to the user

set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE user_service;
    CREATE DATABASE template_service;
    CREATE DATABASE email_service;
    CREATE DATABASE push_service;
    CREATE DATABASE auth_service;
    GRANT ALL PRIVILEGES ON DATABASE user_service TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE template_service TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE email_service TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE push_service TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE auth_service TO $POSTGRES_USER;
EOSQL