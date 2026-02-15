const fs = require('fs');

const file = fs.readFileSync('./doom.s', 'utf-8');
const lines = file.trim().split('\n');

const index = 81

let f = []
const pre = '; function body '
let dentro = false

for (const line of lines) {
    if (line.startsWith(pre + index)) {
        dentro = true
    }
    else if (dentro) {
        if (line.startsWith(pre)) {
            break
        }
        f.push(line)
    }
}

console.log(f.join('\n'))