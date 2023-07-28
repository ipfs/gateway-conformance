const fs = require("fs");

// # read json from stdin:
let lines = fs.readFileSync(0, "utf-8");
lines = JSON.parse(lines);

// # clean input
lines = lines.filter((line) => {
  const { Test } = line;
  return Test !== undefined;
});

lines = lines.filter((line) => {
  const { Action } = line;
  return ["pass", "fail", "skip"].includes(Action);
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

const depth = process.argv[2] && parseInt(process.argv[2], 10) || 1;

// test result is a map { [path_str]: { [path], [action]: count } }
const testResults = {};

lines.forEach((line) => {
  const { Path, Action } = line;
  let current = testResults;

  const path = Path.slice(0, depth)
  const key = path.join(" > ");

  if (!current[key]) {
    current[key] = {Path: path, "pass": 0, "fail": 0, "skip": 0, "total": 0};
  }
  current = current[key];

  current[Action] += 1;
  current["total"] += 1;
});

// output result to stdout
fs.writeFileSync(1, JSON.stringify(testResults, null, 2));
