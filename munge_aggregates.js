/**
 * This script takes the sqlite database from the munging process
 * and generates Hugo content files.
 */
const sqlite3 = require('sqlite3').verbose();
const fs = require('fs');
const util = require('util');
const path = require('path');
const matter = require('gray-matter');

// first parameter: the input database file
// second parameter: the output www/ folder
const dbFile = process.argv[2];
const hugoOutput = process.argv[3];

if (!dbFile || !hugoOutput) {
    console.error("Usage: node munge_sql.js <input.db> <output>");
    process.exit(1);
}

/**
 * @param {string} u a spec URL like "specs.ipfs.tech/http-gateways/path-gateway/#if-none-match-request-header"
 * @returns the spec's parent, or "null" if it's a top-level spec
 */
const computeParent = (u) => {
    const url = new URL(u);
    const segments = url.pathname.split('/').filter(Boolean);

    // if there's a hash, consider it as a segment
    if (url.hash) segments.push(url.hash.substring(1));

    if (segments.length <= 1) {
        return "null";
    }

    const parent = segments.slice(0, -1).join('/');
    return `${url.protocol}//${url.host}/${parent}`
};

/**
 * @param {string} u a spec URL like "specs.ipfs.tech/http-gateways/path-gateway/#if-none-match-request-header"
 * @returns the spec's name, or the hash if it's a top-level spec and whether it was found in a hash
 */
const computeName = (u) => {
    const url = new URL(u);

    if (url.hash) {
        return {
            isHashed: true,
            name: url.hash.substring(1),
        };
    }

    const segments = url.pathname.split('/').filter(Boolean);

    if (segments.length === 0) {
        throw new Error(`Invalid spec URL: ${u}`);
    }

    return {
        isHashed: false,
        name: segments[segments.length - 1],
    };
};

const getTestRunDetails = async (jobUrl) => {
    if (!jobUrl) {
        return {};
    }

    const match = jobUrl.match(/https:\/\/github\.com\/([^\/]+)\/([^\/]+)\/actions\/runs\/(\d+)/);
    if (!match) {
        console.warn('Invalid URL format:', jobUrl);
        return {};
    }

    const [, owner, repo, run_id] = match;

    const apiUrl = `https://api.github.com/repos/${owner}/${repo}/actions/runs/${run_id}`;

    try {
        // Required node18
        const result = await fetch(apiUrl, {
            headers: {
                'Accept': 'application/vnd.github.v3+json'
            }
        })
        const { created_at, head_sha, head_branch, run_started_at } = await result.json()
        return { created_at, head_sha, head_branch, run_started_at }
    } catch (e) {
        console.error(`Error fetch ${jobUrl} details:`, e);
        return {}
    }
}

const main = async () => {
    let db = new sqlite3.Database(dbFile, (err) => {
        if (err) {
            console.error(err.message);
        }
        console.log('Connected to the database.');
    });
    const all = util.promisify(db.all.bind(db));

    // Query to fetch all test runs
    const implementationsQuery = `
        SELECT implementation_id AS id, version, time, job_url
        FROM TestRun
        ORDER BY implementation_id, version, time;
    `;
    const allRuns = await all(implementationsQuery);

    const runs = {};
    for (const row of allRuns) {
        const { id, version, ...rest } = row;
        if (!runs[id]) {
            runs[id] = {};
        }
        const testRunDetails = await getTestRunDetails(rest.job_url);

        runs[id][version] = { ...rest, ...testRunDetails };
    }
    outputJSON("data/testruns.json", runs);

    // Query to fetch all test groups
    const testsQuery = `
        SELECT
            full_name,
            name,
            parent_test_full_name,
            GROUP_CONCAT(DISTINCT test_run_version) AS versions
        FROM TestResult
        GROUP BY full_name, name
        ORDER BY name
    `;
    const testsRows = await all(testsQuery);
    const groups = {};
    const flatTestGroups = {}; // used for specs generation.

    for (const row of testsRows) {
        const { versions, full_name, name, parent_test_full_name } = row;
        const slug = slugifyTestName(full_name);

        if (!groups[parent_test_full_name]) {
            groups[parent_test_full_name] = {};
        }

        const g = { versions: versions?.split(',') || [], name, full_name, slug };

        groups[parent_test_full_name][full_name] = g;
        flatTestGroups[full_name] = g;
    }
    outputJSON("data/testgroups.json", groups);

    // Query to fetch all test specs
    const specsQuery = `
        SELECT
            spec_url as full_name,
            GROUP_CONCAT(DISTINCT test_run_version) AS versions
        FROM TestSpecs
        GROUP BY full_name
        ORDER BY full_name
    `;
    const specsRows = await all(specsQuery);
    const specs = {};
    const flatSpecs = {};

    for (const row of specsRows) {
        const { versions, full_name } = row;
        let current = full_name;

        while (current !== "null") {
            const slug = slugify(current);
            const parent = computeParent(current);
            const { name, isHashed } = computeName(current)

            if (!specs[parent]) {
                specs[parent] = {};
            }

            flatSpecs[current] = true

            specs[parent][current] = {
                versions: versions?.split(',') || [],
                spec_full_name: current,
                slug,
                name,
                isHashed,
            };

            current = parent;
        }
    }
    outputJSON("data/specs.json", specs);

    const descendTheSpecsTree = (current, path) => {
        Object.entries(specs[current] || {})
            .forEach(([key, spec]) => {
                const addSpecs = (current) => {
                    let hashes = [...(current.specs || []), spec.name];
                    hashes = [...new Set(hashes)]; // deduplicate
                    return { ...current, hashes }
                };

                // To reproduce the structure of URLs and hashes, we update existing specs pages
                if (spec.isHashed) {
                    const p = path.join("/");
                    outputFrontmatter(
                        `content/specs/${p}/_index.md`,
                        addSpecs
                    );
                    // We assume there are no recursion / children for hashes
                    return
                }

                const newPath = [...path, spec.name];
                const p = newPath.join("/");

                outputFrontmatter(`content/specs/${p}/_index.md`, {
                    ...spec,
                    title: spec.name,
                });

                descendTheSpecsTree(key, newPath);
            })
    }

    descendTheSpecsTree("null", [])

    // Aggregate test results per specs
    const specsTestGroups = {};

    for (const fullName of Object.keys(flatSpecs)) {
        // list all the test names for a given spec.
        // we prefix search the database for spec_urls starting with the spec name
        const specsQuery = `
            SELECT
                test_full_name
            FROM TestSpecs
            WHERE spec_url LIKE ?
            ORDER BY test_full_name
        `;
        const tests = await all(specsQuery, [fullName + '%']);

        const s = tests.map(x => x.test_full_name)
            .reduce((acc, name) => {
                return {
                    ...acc,
                    [name]: flatTestGroups[name]
                }
            }, {});
        specsTestGroups[fullName] = s;
    }

    outputJSON("data/specsgroups.json", specsTestGroups);

    // Query to fetch all stdouts
    const logsQuery = `
        SELECT
            test_run_implementation_id AS implementation_id,
            test_run_version AS version,
            test_full_name AS full_name,
            stdout
        FROM TestLog
        ORDER BY test_full_name
    `;
    const logsRow = await all(logsQuery);
    const logs = {};
    for (const row of logsRow) {
        const { implementation_id, version, full_name, stdout } = row;

        if (!logs[implementation_id]) {
            logs[implementation_id] = {};
        }

        if (!logs[implementation_id][version]) {
            logs[implementation_id][version] = {};
        }

        logs[implementation_id][version][full_name] = stdout;
    }
    outputJSON("data/testlogs.json", logs);

    // Generate test results for every run
    for ({ id, version } of allRuns) {
        const testResultQuery = `
            WITH LeafTests AS (
                -- Identify leaf tests (tests without a descendant)
                SELECT full_name, outcome
                FROM TestResult tr1
                WHERE test_run_implementation_id = ? AND test_run_version = ?
                AND NOT EXISTS (
                    SELECT 1
                    FROM TestResult tr2
                    WHERE tr2.test_run_implementation_id = tr1.test_run_implementation_id
                        AND tr2.test_run_version = tr1.test_run_version
                        AND tr2.parent_test_full_name = tr1.full_name
                )
            )

            SELECT
                tr.full_name,
                tr.name,
                tr.parent_test_full_name,
                COUNT(CASE WHEN lt.outcome = 'pass' THEN 1 ELSE NULL END) AS passed_leave,
                COUNT(CASE WHEN lt.outcome = 'fail' THEN 1 ELSE NULL END) AS failed_leaves,
                COUNT(CASE WHEN lt.outcome = 'skip' THEN 1 ELSE NULL END) AS skipped_leaves,
                COUNT(lt.full_name) AS total_leaves
            FROM TestResult tr
            LEFT JOIN LeafTests lt
                ON lt.full_name LIKE tr.full_name || '%'
            WHERE tr.test_run_implementation_id = ? AND tr.test_run_version = ?
            GROUP BY tr.full_name
            ORDER BY tr.full_name;
        `;
        const rows = await all(testResultQuery, [id, version, id, version]);

        const testResults = {};
        for (const row of rows) {
            testResults[row.full_name] = { ...row, slug: slugifyTestName(row.full_name) };
        }
        outputJSON(`data/testresults/${id}/${version}.json`, testResults);
    }

    // Generate Test taxonomies
    // List all the tests full names.
    const testsTaxonomyQuery = `
        SELECT DISTINCT
            tr.full_name,
            tr.name,
            tr.test_run_version,
            tm.key,
            tm.value
        FROM TestResult tr
        LEFT JOIN TestMetadata tm
            ON tm.test_run_implementation_id = tr.test_run_implementation_id
            AND tm.test_run_version = tr.test_run_version
            AND tm.test_full_name = tr.full_name
        ORDER BY full_name
    `;
    const testsTaxonomyRows = await all(testsTaxonomyQuery);

    const testsTaxonomy = {};
    for (const row of testsTaxonomyRows) {
        const { full_name, test_run_implementation_id, test_run_version } = row;
        const slug = slugifyTestName(full_name);
        const name = decodeURIComponent(row.name);

        if (!testsTaxonomy[full_name]) {
            testsTaxonomy[full_name] = {
                slug,
                name,
                full_name,
                versions: [],
            };
        }

        addUniq(testsTaxonomy[full_name].versions, test_run_version);

        if (row.key !== null) {
            const key = row.key + 's'; // taxonomies are plural, ipip => ipips

            if (!testsTaxonomy[full_name][key]) {
                testsTaxonomy[full_name][key] = [];
            }
            const value = JSON.parse(row.value);
            addUniq(testsTaxonomy[full_name][key], value);
        }
    }

    for (const test of Object.values(testsTaxonomy)) {
        outputFrontmatter(`content/tests/${test.slug}/_index.md`, {
            ...test,
            title: test.name
        });
    }

    // Generate Results taxonomies
    // List all the tests implementation / version / tests full names / outcome
    const resultsTaxonomyQuery = `
        SELECT
            test_run_implementation_id AS implementation_id,
            test_run_version AS version,
            full_name,
            name,
            outcome
        FROM TestResult
        ORDER BY test_run_implementation_id, test_run_version, full_name
    `;
    const resultsTaxonomyRows = await all(resultsTaxonomyQuery);

    const resultsTaxonomy = {};
    for (const row of resultsTaxonomyRows) {
        const { implementation_id, version, full_name, outcome } = row;
        const slug = slugifyTestName(full_name);
        const name = decodeURIComponent(row.name);

        if (!resultsTaxonomy[implementation_id]) {
            resultsTaxonomy[implementation_id] = {};
        }

        if (!resultsTaxonomy[implementation_id][version]) {
            resultsTaxonomy[implementation_id][version] = {};
        }

        if (!resultsTaxonomy[implementation_id][version][full_name]) {
            resultsTaxonomy[implementation_id][version][full_name] = {
                slug,
                name,
                full_name,
                outcome,
            };
        }
    }

    for (const [implementation_id, versions] of Object.entries(resultsTaxonomy)) {
        outputFrontmatter(`content/results/${implementation_id}/_index.md`, {
            implementation_id,
            title: implementation_id
        });

        for (const [version, tests] of Object.entries(versions)) {
            outputFrontmatter(`content/results/${implementation_id}/${version}/_index.md`, {
                implementation_id,
                version,
                title: version
            });

            for (const test of Object.values(tests)) {
                outputFrontmatter(`content/results/${implementation_id}/${version}/${test.slug}/_index.md`, {
                    ...test,
                    implementation_id,
                    version,
                    title: test.name
                });
            }
        }
    }

    // Close the database connection when you're done
    db.close((err) => {
        if (err) {
            console.error(err.message);
        }
        console.log('Closed the database connection.');
    });
}

const slugify = (str) => {
    // https://byby.dev/js-slugify-string
    return String(str)
        .normalize('NFKD') // split accented characters into their base characters and diacritical marks
        .replace(/[\u0300-\u036f]/g, '') // remove all the accents, which happen to be all in the \u03xx UNICODE block.
        .trim() // trim leading or trailing whitespace
        .toLowerCase() // convert to lowercase
        .replace(/\s+/g, '_') // replace spaces with underscore
        .replace(/[.,()"/]/g, '-') // remove the characters (, ) and ~
        .replace(/[^a-z0-9 -\/]/g, '-') // remove non-alphanumeric characters
        .replace(/_+/g, '_') // remove consecutive underscores
        .replace(/-+/g, '-') // remove consecutive dashes
}

const slugifyTestName = (str) => {
    let x = String(str).split('/');
    x = x.map(part => {
        // url decode to handle %20, etc
        part = decodeURIComponent(part);
        return slugify(part.replace(/([a-z])([A-Z])/g, '$1-$2') // Convert CamelCase to kebab-case
        )
    })
    return x.join('/');
}

const outputJSON = (p, data) => {
    const json = JSON.stringify(data, null, 2);
    const fullPath = `${hugoOutput}/${p}`;

    const folders = path.dirname(fullPath);
    if (!fs.existsSync(folders)) {
        fs.mkdirSync(folders, { recursive: true });
    }

    fs.writeFileSync(fullPath, json);
}

const outputFrontmatter = (p, dataOrUpdate) => {
    const fullPath = `${hugoOutput}/${p}`;

    // TODO: implement update frontmatter

    const folders = path.dirname(fullPath);
    if (!fs.existsSync(folders)) {
        fs.mkdirSync(folders, { recursive: true });
    }

    // if file exists, load it with gray matter
    const content = {
        content: "",
        data: {}
    }
    if (fs.existsSync(fullPath)) {
        const existing = matter.read(fullPath);
        content.content = existing.content;
        content.data = existing.data;
    }

    if (typeof dataOrUpdate === "function") {
        content.data = dataOrUpdate(content.data);
    } else {
        content.data = { ...content.data, ...dataOrUpdate };
    }

    const md = matter.stringify(content.content, content.data);
    fs.writeFileSync(fullPath, md);
}

const addUniq = (arr, value) => {
    if (!arr.includes(value)) {
        arr.push(value);
    }
}

main()
    .then(() => {
        console.log("done");
    })
    .catch((e) => {
        console.error(e);
        process.exit(1);
    })
