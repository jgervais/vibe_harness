import fs from 'fs';

async function example() {
    const data = fs.readFileSync("file.txt");
    console.log("Read file: file.txt");
}