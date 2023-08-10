const fs = require('fs');

const jsonFilePath = process.argv[2];
const markdownFilePath = process.argv[3];

// Check if file paths are provided
if (!jsonFilePath || !markdownFilePath) {
    console.error('Both a JSON file path and a Markdown file path must be provided.');
    process.exit(1);
}

const jsonData = JSON.parse(fs.readFileSync(jsonFilePath, 'utf8'));
const sortedKeys = Object.keys(jsonData).sort((a, b) => b.length - a.length);

let markdown = fs.readFileSync(markdownFilePath, 'utf8');

for (const key of sortedKeys) {
    const newName = jsonData[key][1]
        ? `[${jsonData[key][0]}](${jsonData[key][1]})`
        : jsonData[key][0];

    const regex = new RegExp(key, 'g');
    markdown = markdown.replace(regex, newName);
}

// output the new markdown to stdout
fs.writeFileSync(1, markdown);