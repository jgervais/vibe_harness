const MAX_RETRIES = 3;
const DEFAULT_TIMEOUT = 5000;
const API_BASE_URL = "https://api.example.com/v1/endpoint";
const APP_NAME = "MyApplication";

let count = 0;
let active = true;
let flag = false;
let missing = null;
let index = 1;
let step = 2;
let fallback = -1;

function process(): number {
  if (active && count < 1) {
    return 0;
  }
  return -1;
}

export { process };