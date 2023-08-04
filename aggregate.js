const fs = require("fs");

// # we group test results by Path, depth is the number of levels to group by
const depth = process.argv[2] && parseInt(process.argv[2], 10) || 1;

// # read json from stdin:
let lines = fs.readFileSync(0, "utf-8");
lines = JSON.parse(lines);

// # clean input
lines = lines.filter((line) => {
  const { Test } = line;
  return Test !== undefined;
});

// # extract test metadata
//   action is output, and starts with ".* --- META: (.*)"
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

lines = lines.map((line) => {
  const metadata = extractMetadata(line);

  if (!metadata) {
    return line;
  }

  return {
    ...line,
    Action: "meta",
    Metadata: metadata,
  }
});

// # keep the test result lines and metadata only
lines = lines.filter((line) => {
  const { Action } = line;
  return ["pass", "fail", "skip", "meta"].includes(Action);
});

// # add "Path" field by parsing "Name" and split by "/"
//   also update the name to make it readable
//   also remove "Time" field while we're at it
lines = lines.map((line) => {
  const { Test, Time, ...rest } = line;
  const path = Test.split("/").map((name) => {
    return name.replace(/_/g, " ");
  });

  return { ...rest, Path: path };
});

// # Aggregate all known "Path" values, use a tree structure to represent it
//   {
//       child1: {
//           child2: {
//               ...,
//           }
//       }
//   }
const testTree = {};

lines.forEach((line) => {
  const { Path } = line;
  let current = testTree;

  Path.forEach((path) => {
    if (!current[path]) {
      current[path] = {};
    }
    current = current[path];
  });
})

// prepare metadata up the tree
const metadataTree = {};

// sort lines so that the one with the longest path is processed first
const sortedLines = lines.sort((a, b) => {
  return b.Path.length - a.Path.length;
});

sortedLines.forEach((line) => {
  const { Path, Action, Metadata } = line;
  let current = metadataTree;

  if (Action !== "meta") {
    return;
  }

  Path.forEach((path) => {
    if (!current[path]) {
      current[path] = {};
    }
    current = current[path];
    current["meta"] = { ...current["meta"], ...Metadata };
  });
});

const getMetadata = (path) => {
  let current = metadataTree;

  path.forEach((path) => {
    if (!current[path]) {
      return null;
    }
    current = current[path];
  });

  return current["meta"];
}

// # Drop all lines where the Test "Path" does not point to a leaf
//   if the test has children then we don't really care about it's pass / fail / skip status,
//   we'll aggregate its children results'
lines = lines.filter((line) => {
  const { Path } = line;
  let current = testTree;

  Path.forEach((path) => {
    if (!current[path]) {
      return false;
    }
    current = current[path];
  });

  // if current has children, it is not a leaf
  return Object.keys(current).length === 0;
});

// # Aggregate by Path and count actions

// test result is a map { [path_str]: { [path], [action]: count } }
const testResults = {};

lines.forEach((line) => {
  const { Path, Action } = line;
  let current = testResults;

  const path = Path.slice(0, depth)
  const key = path.join(" > ");

  if (!current[key]) {
    current[key] = { Path: path, "pass": 0, "fail": 0, "skip": 0, "total": 0, "meta": getMetadata(path) || {} };
  }
  current = current[key];

  if (Action === "meta") {
    const { Metadata } = line;
    current["meta"] = { ...current["meta"], ...Metadata };
    return;
  } else {
    current[Action] += 1;
    current["total"] += 1;
  }
});

// output result to stdout
fs.writeFileSync(1, JSON.stringify(testResults, null, 2));
