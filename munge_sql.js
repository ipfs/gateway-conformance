/**
 * This script takes the munged JSON files and puts them into a SQLite database.
 */
const sqlite3 = require('sqlite3').verbose();
const fs = require('fs');
const util = require('util');

// first parameter: the output database file
// all the following parameters: the input json files
const dbFile = process.argv[2];
const files = process.argv.slice(3);

if (!dbFile || files.length === 0) {
    console.error("Usage: node munge_sql.js <output.db> <input1.json> <input2.json> ...");
    process.exit(1);
}

const main = async () => {
    let db = new sqlite3.Database(dbFile, (err) => {
        if (err) {
            console.error(err.message);
        }
        console.log('Connected to the database.');
    });

    const run = util.promisify(db.run.bind(db));

    // Create the TestResult table if it doesn't exist
    await run(`
        CREATE TABLE IF NOT EXISTS TestRun (
            implementation_id TEXT,
            version TEXT,
            time DATETIME,
            job_url TEXT,

            PRIMARY KEY (implementation_id, version)
        );
    `);

    await run(`
        CREATE TABLE IF NOT EXISTS TestResult (
            test_run_implementation_id TEXT,
            test_run_version TEXT,

            full_name TEXT,
            name TEXT,
            outcome TEXT CHECK(outcome IN ('pass', 'fail', 'skip')),

            parent_test_full_name TEXT,

            PRIMARY KEY (test_run_implementation_id, test_run_version, full_name),

            -- parent hierarchy
            FOREIGN KEY (test_run_implementation_id, test_run_version, parent_test_full_name)
                REFERENCES TestResult (test_run_implementation_id, test_run_version, full_name),
            
            -- test run
            FOREIGN KEY (test_run_implementation_id, test_run_version)
                REFERENCES TestRun (implementation_id, version)
        );
    `);

    // TODO: verify me.
    await run(`
        CREATE TABLE IF NOT EXISTS TestMetadata (
            test_run_implementation_id TEXT,
            test_run_version TEXT,
            test_full_name TEXT,

            key TEXT,
            value JSON,

            PRIMARY KEY (test_run_implementation_id, test_run_version, test_full_name, key),

            -- test run
            FOREIGN KEY (test_run_implementation_id, test_run_version)
                REFERENCES TestRun (implementation_id, version)

            -- test result
            FOREIGN KEY (test_run_implementation_id, test_run_version, test_full_name)
                REFERENCES TestResult (test_run_implementation_id, test_run_version, full_name)
        );
    `)

    await run(`
        CREATE TABLE IF NOT EXISTS TestLog (
            test_run_implementation_id TEXT,
            test_run_version TEXT,
            test_full_name TEXT,

            stdout TEXT,

            -- test run
            FOREIGN KEY (test_run_implementation_id, test_run_version)
                REFERENCES TestRun (implementation_id, version)

            -- test result
            FOREIGN KEY (test_run_implementation_id, test_run_version, test_full_name)
                REFERENCES TestResult (test_run_implementation_id, test_run_version, full_name)
        );
    `)

    for (const file of files) {
        const fileName = file.split("/").slice(-1)[0].split(".")[0];
        const implemId = fileName;

        const content = JSON.parse(fs.readFileSync(file));
        const { TestMetadata, ...tests } = content;

        const time = TestMetadata?.time;
        const { version, job_url } = TestMetadata?.meta || {};

        await run(`
            INSERT INTO TestRun (implementation_id, version, time, job_url)
            VALUES (?, ?, ?, ?)
            ON CONFLICT (implementation_id, version) DO UPDATE SET
                time = excluded.time,
                job_url = excluded.job_url
        `, [implemId, version, time, job_url]);

        // process all the tests. Start with the roots.
        const sorted = Object.keys(tests).sort();

        for (testId of sorted) {
            const test = tests[testId];

            const fullName = testId
            const name = test.path[test.path.length - 1];
            const outcome = test.outcome;
            const parentFullName = testId.split('/').slice(0, -1).join("/") || null;

            await run(`
                INSERT INTO TestResult (test_run_implementation_id, test_run_version, full_name, name, outcome, parent_test_full_name)
                VALUES (?, ?, ?, ?, ?, ?)
            `, [implemId, version, fullName, name, outcome, parentFullName]);

            for (const [key, value] of Object.entries(test.meta ?? {})) {
                await run(`
                    INSERT INTO TestMetadata (test_run_implementation_id, test_run_version, test_full_name, key, value)
                    VALUES (?, ?, ?, ?, ?)
                `, [implemId, version, fullName, key, JSON.stringify(value)]);
            }

            await run(`
                INSERT INTO TestLog (test_run_implementation_id, test_run_version, test_full_name, stdout)
                VALUES (?, ?, ?, ?)
            `, [implemId, version, fullName, test.output]);

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

main()
    .then(() => {
        console.log("done");
    })
    .catch((e) => {
        console.error(e);
        process.exit(1);
    })