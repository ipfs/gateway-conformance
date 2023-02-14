import { execSync } from "child_process";
import {
  exportFixtureDefinitionToTs,
  listFixtures,
  loadFixturesDefinition,
  loadFixtureYaml,
  generateFixturesCarFile,
} from "./util/fixtures.js";

async function prepare(...args: string[]) {
  const cids = new Set<string>();

  for (const inputPath of listFixtures()) {
    console.log(`Fixture: ${inputPath}`);
    const fixture = await loadFixtureYaml(inputPath);
    const fixtureDefinition = await loadFixturesDefinition(fixture);

    const outputPath = inputPath.replace("fixtures.yaml", "fixtures.ts");
    exportFixtureDefinitionToTs(outputPath, fixtureDefinition);

    // Aggregate all known CIDs to generate our test car file
    Object.values(fixture.ipfs).forEach((cid) => cids.add(cid));
  }

  generateFixturesCarFile("fixtures.car", cids);
}

prepare(...process.argv.slice(2))
  .then(() => process.exit(0))
  .catch((err) => {
    console.error(err);
    process.exit(1);
  });
