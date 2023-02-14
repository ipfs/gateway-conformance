import fs from "fs";
import path from "path";
import { parse } from "yaml";
import { dirname } from "path";
import { fileURLToPath } from "url";
import { execSync } from "child_process";

const __dirname = path.join(dirname(fileURLToPath(import.meta.url)), "..");

const testFolder = path.join(__dirname, "test");

interface FixturesYaml {
  ipfs: { [key: string]: string };
}

interface IFixturesDefinition {
  ipfs: {
    [key: string]: IPFSFixture;
  };
}

interface IPFSFixture {
  [key: string]: IPFSFixture | string;
  _cid: string;
  _block: string;
  _dag: string;
}

export function* listFixtures(
  dir: string = testFolder,
  suffix: string = "fixtures.yaml"
): Generator<string> {
  const files = fs.readdirSync(dir);

  for (const file of files) {
    const filePath = path.join(dir, file);
    const isDirectory = fs.statSync(filePath).isDirectory();

    if (isDirectory) {
      yield* listFixtures(filePath, suffix);
    } else {
      if (file.endsWith(suffix)) {
        yield filePath;
      }
    }
  }
}

export async function loadFixtureYaml(path: string): Promise<FixturesYaml> {
  // TODO: validate data
  return parse(fs.readFileSync(path).toString("utf-8"));
}

export async function loadFixturesDefinition(
  yaml: FixturesYaml
): Promise<IFixturesDefinition> {
  const structure: IFixturesDefinition = { ipfs: {} };

  for (const [name, cid] of Object.entries(yaml.ipfs)) {
    console.log(`${name}: ${cid}`);
    structure.ipfs[name] = await loadIPFSFixture(cid);
  }

  return structure;
}

async function loadIPFSFixture(cid: string): Promise<IPFSFixture> {
  const blockData = execSync(`ipfs block get ${cid}`).toString("base64");
  const dagData = execSync(`ipfs dag export ${cid}`).toString("base64");

  const result: IPFSFixture = {
    _cid: cid,
    _block: blockData,
    _dag: dagData,
  };

  try {
    const out = execSync(`ipfs ls ${cid}`);
    const lines = out
      .toString("utf-8")
      .split("\n")
      .filter((line) => !!line);

    if (lines.length === 0) {
      return result;
    }

    for (const line of lines) {
      const [cid, _size, name] = line.split(/\s+/);
      const cleanName = name.replace(/\/$/, "");

      if (name === "_cid") {
        throw new Error(`collision with names`);
      }

      result[cleanName] = await loadIPFSFixture(cid);
    }
  } catch (err) {
    console.error(err);
  }

  return result;
}

export function exportFixtureDefinitionToTs(
  outputPath: string,
  structure: IFixturesDefinition
) {
  const output = `
// This file was generated from the fixtures.yaml file.    
const fixture = ${JSON.stringify(structure, null, 2)}

export const rawBlock = (x: { _block: string }): Buffer => {
  return Buffer.from(x._block, "base64");
};

export const blockSize = (x: { _block: string }): number => {
  return rawBlock(x).length;
};

export const blockAsString = (x: { _block: string }): string => {
  return rawBlock(x).toString("utf-8");
};

export const rawDag = (x: { _dag: string }): Buffer => {
  return Buffer.from(x._dag, "base64");
};

export const dagAsString = (x: { _dag: string }): string => {
  return rawDag(x).toString("utf-8");
};

export const dagSize = (x: { _dag: string }): number => {
  return rawDag(x).length;
};

export default fixture.ipfs
`;

  fs.writeFileSync(outputPath, output);
}

export function generateFixturesCarFile(outputPath: string, cids: Set<string>) {
  // Now go through every known CIDs and export them into a single fixtures.car file.

  // TODO: this is a naive way to merge CIDs / car files.
  const newName = `test-${Date.now()}`;
  execSync(`ipfs files mkdir /${newName}`);

  for (const cid of cids) {
    console.log(`Importing ${cid} into MFS`);
    const out = execSync(`ipfs files cp /ipfs/${cid} /${newName}/${cid}`);
  }

  const out = execSync(`ipfs files stat --hash /${newName}`);
  const hash = out.toString("utf-8").trim();

  console.log(`Exporting MFS folder to car file: ${hash}`);
  const out2 = execSync(`ipfs dag export ${hash} > ${outputPath}`);
  console.log(out2.toString());
}
