const fs = require("fs");

const TestMetadata = "TestMetadata";
const METADATA_TEST_GROUP = "group";

// retrieve the list of input files from the command line
const files = process.argv.slice(2);

// read all input files (json)
const inputs = files.map((file) => {
    return JSON.parse(fs.readFileSync(file, 'utf8'));
});

// merge all the unique keys & metadata from all the inputs
const metadata = {}
inputs.forEach((input) => {
    Object.keys(input).forEach((key) => {
        metadata[key] = { ...metadata[key], ...input[key]["meta"] || {} };
    });
});
delete metadata[TestMetadata]; // Extract TestMetadata which is a special case

// generate groups: an array of {group, key} objects
// where group is the group name (or undefined), and key is the test key name (or undefined)
// It represents the table leftmost column.
//
// Group1
//  Group1 - Test1
//  Group1 - Test2
// Group2
// ...
const groups = []
const groupsAdded = new Set();
Object.entries(metadata).forEach(([key, value]) => {
    const group = value[METADATA_TEST_GROUP] || undefined;

    if (!groupsAdded.has(group)) {
        groups.push({ group, key: undefined });
        groupsAdded.add(group);
    }

    groups.push({ group, key });
});

// sort the groups so that the tests are ordered by group, then by key.
// undefined groups are always at the end.
groups.sort((a, b) => {
    if (a.group === b.group) {
        if (a.key === undefined) {
            return -1;
        }
        if (b.key === undefined) {
            return 1;
        }
        return a.key.localeCompare(b.key);
    }

    if (a.group === undefined) {
        return 1;
    }

    if (b.group === undefined) {
        return -1;
    }

    return a.group.localeCompare(b.group);
});

// generate a table
const columns = [];

// add the leading column ("gateway", "version", "group1", "test11", ... "test42")
const leading = ["gateway", "version"];
groups.forEach(({ group, key }) => {
    if (key === undefined) {
        leading.push(`**${group || 'Other'}**`);
        return;
    }

    // Skip the "Test" prefix
    let niceKey = key.replace(/^Test/, '');

    const m = metadata[key];
    if (m.specs && m.specs.length > 0) {
        if (m.specs.length === 1) {
            niceKey = `[${niceKey}](https://${m.specs[0]})`
        } else {
            const urls = m.specs.map((url, index) => `[${index}](https://${url})`);
            niceKey = `${niceKey} (${urls.join(', ')})`;
        }
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
    let version = 'unknown';
    let jobURL = null;

    if (input[TestMetadata]) {
        const metadata = input[TestMetadata]["meta"];
        version = metadata['version'];
        jobURL = metadata['job_url'];
    }

    let versionCell = version
    if (jobURL) {
        versionCell = `[${version}](${jobURL})`;
    }

    const col = [name, versionCell];

    // extract results
    groups.forEach(({ group, key }) => {
        if (key === undefined) {
            col.push(null);
            return;
        }
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
