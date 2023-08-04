const fs = require("fs");

const TestMetadata = "TestMetadata";
const METADATA_TEST_GROUP = "group";
const METADATA_IPIP = "ipip";

// retrieve the list of input files from the command line
const files = process.argv.slice(2);

// read all input files (json)
const inputs = files.map((file) => {
    return JSON.parse(fs.readFileSync(file, 'utf8'));
});

// merge all the unique keys & metadata from all the inputs
let keys = new Set();
const metadata = {}
inputs.forEach((input) => {
    Object.keys(input).forEach((key) => {
        keys.add(key);

        // add metadata keys (do not care about different metadatas for now)
        metadata[key] = { ...metadata[key], ...input[key]["meta"] || {} };
    });
});

keys.delete(TestMetadata); // Extract TestMetadata which is a special case
keys = Array.from(keys).sort();

// generate a table
const columns = [];

// add the leading column ("gateway", "version", "key1", "key2", ... "keyN")
const leading = ["gateway", "version"];
keys.forEach((key) => {
    const m = metadata[key];

    // Skip the "Test" prefix
    let niceKey = key.replace(/^Test/, '');

    niceKey = m[METADATA_TEST_GROUP] || niceKey;

    // Add ipip link if available
    if (m[METADATA_IPIP]) {
        niceKey = `[${niceKey}](https://specs.ipfs.tech/ipips/${m[METADATA_IPIP]})`;
    }

    leading.push(niceKey);
});
columns.push(leading);

// add the data for every input
const cellRender = (cell) => {
    if (cell === null) {
        return '';
    }

    if (cell['fail'] > 0) {
        return `:red_circle: (${cell['pass']} / ${cell['total']})`;
    }
    if (cell['skip'] > 0) {
        return `:yellow_circle: (skipped)`;
    }
    if (cell['pass'] > 0) {
        return `:green_circle: (${cell['pass']} / ${cell['total']})`;
    }

    throw new Error(`Unhandled cell value: ${JSON.stringify(cell)}`);
}

inputs.forEach((input, index) => {
    // clean name (remove path and extension)
    let name = files[index].replace(/\.json$/, '').replace(/^.*\//, '');

    // extract TestMetadata & version
    const metadata = input[TestMetadata]["meta"];
    const version = metadata['version'];

    const col = [name, version];

    // extract results
    keys.forEach((key) => {
        col.push(cellRender(input[key] || null));
    });
    columns.push(col);
});

// # Rotate the table
// it's easier to create the table by column, but we want to render it by row
let rows = columns[0].map((_, i) => columns.map(col => col[i]));

// # Render the table into a markdown table

// add the hyphen header row after the first row
const hyphenated = rows[0].map((x, i) => {
    if (i === 0) {
        return '-'.repeat(Math.max(0, x.length - 2)) + '-:'
    }
    return ':-' + '-'.repeat(Math.max(0, x.length - 2));
})

rows = [
    rows[0],
    hyphenated,
    ...rows.slice(1),
]

let markdown = rows.map(row => '| ' + row.join(' | ') + ' |').join('\n');

// output the table to stdout
fs.writeFileSync(1, markdown);
