/**
 * This file loads a test2json output from stdin,
 * and generate a test result structure.
 * 
 * The output is a map of test names to test results.
 * 
 * A test result is an object with the following fields:
 * - path: the test path (["Test Something", "Sub Test", ...])
 * - output: the test stdout
 * - outcome: "pass" | "fail" | "skip" | "unknown"
 * - time: the test finish time
 * - meta: test metadata such as "version", "ipip", etc.
 */
const fs = require("fs");

// # read jsonlines from stdin:
let lines = fs.readFileSync(0, "utf-8");

// # clean input 
lines = lines
    .split("\n")
    // ## remove empty lines
    .filter((line) => line !== "")
    // ## extract json
    .map((line) => {
        try {
            return JSON.parse(line);
        } catch (e) {
            throw new Error(`Failed to parse line: ${line}: ${e}`);
        }
    })
    // ## Drop lines that are not about a Test
    .filter((line) => {
        const { Test } = line;
        return Test !== undefined;
    });

// # extract test metadata
//   For now we look for the '--- META:' marker in test logs.
//   see details in https://github.com/ipfs/gateway-conformance/pull/125
const extractMetadata = (line) => {
    const { Action, Output } = line;

    if (Action !== "output") {
        return null;
    }

    const match = Output.match(/.* --- META: (.*)/);

    if (!match) {
        return null;
    }

    const metadata = match[1];
    return JSON.parse(metadata);
}

// # Group all lines by Test name, and Action
const groups = {};
lines.forEach((line) => {
    const { Test, Action } = line;

    if (!Test || !Action) {
        throw new Error(`Missing Test field in line: ${JSON.stringify(line)}`);
    }

    if (!groups[Test]) {
        groups[Test] = {};
    }

    if (!groups[Test][Action]) {
        groups[Test][Action] = [];
    }

    groups[Test][Action].push(line);

    // Add metadata while we're at it
    const metadata = extractMetadata(line);

    if (metadata) {
        if (!groups[Test]["meta"]) {
            groups[Test]["meta"] = [];
        }
        groups[Test]["meta"].push(metadata);
    }
});

// # Now that we grouped test results,
//   merge test results into logical aggregates, like stdouts, metadata, etc.
const groupTest = (test) => {
    const { run, output, pass, fail, skip, meta } = test;

    const path = run[0]["Test"].split("/").map((name) => {
        return name.replace(/_/g, " ");
    });

    const outputMerged = output.reduce((acc, line) => {
        const { Output } = line;
        return acc + Output;
    }, "");

    const metaMerged = meta ? meta.reduce((acc, line) => {
        // fail in case of duplicate keys
        if (Object.keys(acc).some((key) => line[key] !== undefined)) {
            throw new Error(`Duplicate metadata key: ${JSON.stringify(line)}`);
        }
        return { ...acc, ...line };
    }) : undefined;

    const outcomeLine = (pass || fail || skip || [{ Action: "Unknown" }])[0];
    const time = outcomeLine["Time"];

    return {
        path,
        output: outputMerged,
        outcome: outcomeLine["Action"],
        time,
        meta: metaMerged
    }
}

const merged = {}
Object.entries(groups).forEach(([test, group]) => {
    merged[test] = groupTest(group);
})

// output result to stdout
const result = merged
fs.writeFileSync(1, JSON.stringify(result, null, 2));