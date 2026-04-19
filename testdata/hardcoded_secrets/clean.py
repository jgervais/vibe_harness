import os

APP_NAME = "myapp"
MAX_RETRIES = 3
DEBUG_MODE = True
PORT = 8080

def connect():
    host = os.environ.get("DB_HOST", "localhost")
    port = os.environ.get("DB_PORT", "5432")
    return f"host={host} port={port}"

def get_config():
    return {
        "name": APP_NAME,
        "version": "1.0.0",
    }