import * as dagPB from "@ipld/dag-pb";
import { MemoryBlockstore } from "blockstore-core/memory";
import * as fs from "fs";
import { getAllFilesSync } from "get-all-files";
import { exporter, UnixFSEntry } from "ipfs-unixfs-exporter";
import {
  ImportCandidate,
  importer,
  ImportResult,
  UserImporterOptions,
} from "ipfs-unixfs-importer";
import { CID } from "multiformats/cid";
import { join as joinPath } from "path";
import { IPNSFixtureOptions, provisionIPNSWithKubo } from "./provisioners.js";

interface FixtureOptions extends UserImporterOptions {}

interface Entry {
  imported: ImportResult;
  exported: UnixFSEntry;
  raw: Buffer | Uint8Array;
  cid: CID;
}

const fixtures: Fixture[] = [];

export class Fixture {
  public readonly path: string;
  public readonly options: FixtureOptions;
  public readonly entries: Entry[];

  constructor(path: string, options: FixtureOptions, entries: Entry[]) {
    this.path = path;
    this.options = options;
    this.entries = entries;

    fixtures.push(this);
  }

  static get(path: string): Fixture {
    const fixture = fixtures.find((fixture) => fixture.path === path);

    if (!fixture) {
      throw new Error(`Fixture ${path} not found`);
    }

    return fixture;
  }

  static getAll(): Fixture[] {
    return fixtures;
  }

  static getAbsolutePath(path: string): string {
    return new URL(joinPath("..", "fixtures", path), import.meta.url).pathname;
  }

  getAbsolutePath(): string {
    return Fixture.getAbsolutePath(this.path);
  }

  static isDirectory(path: string): boolean {
    return fs.lstatSync(Fixture.getAbsolutePath(path)).isDirectory();
  }

  isDirectory(): boolean {
    return Fixture.isDirectory(this.path);
  }

  static async fromPath(path: string, options: FixtureOptions = {}) {
    const absolute = Fixture.getAbsolutePath(path);
    const source: ImportCandidate[] = [];

    if (Fixture.isDirectory(path)) {
      for (const file of getAllFilesSync(absolute)) {
        source.push({
          path: `${path}/${file.slice(`${absolute}/`.length)}`,
          content: fs.readFileSync(file),
        });
      }
    } else {
      source.push({
        path,
        content: fs.readFileSync(absolute),
      });
    }

    const blockstore = new MemoryBlockstore();
    const entries: Entry[] = [];
    for await (const imported of importer(source, blockstore, options)) {
      const exported = await exporter(imported.cid, blockstore);
      let raw;

      if (exported.type === "raw") {
        raw = exported.node;
      } else {
        // @ts-ignore: fix the UInt8Array | PBNode type
        raw = Buffer.from(dagPB.encode(exported.node));
      }
      // TODO: this api deserve to be unified behind a single Fixture object.
      entries.push({
        imported,
        exported,
        raw,
        cid: imported.cid,
      });
    }

    return new Fixture(path, options, entries);
  }

  get(path: string): Entry {
    const entry = this.entries.find((entry) => entry.imported.path === path);

    if (!entry) {
      throw new Error(`Entry not found for path: ${path}`);
    }

    return entry;
  }

  get root(): Entry {
    return this.get("");
  }

  getCID(path: string): CID {
    return this.get(path).imported?.cid;
  }

  get cid(): CID {
    return this.getCID("");
  }

  getRaw(path: string): Buffer | Uint8Array {
    return this.get(path).raw;
  }

  getString(path: string): string {
    return this.getRaw(path)?.toString();
  }

  getLength(path: string): number {
    return this.getRaw(path)?.length;
  }
}

await Promise.all([
  Fixture.fromPath("dir", {
    cidVersion: 1,
    rawLeaves: true,
    wrapWithDirectory: true,
  }),
  Fixture.fromPath("root2", {
    cidVersion: 1,
    rawLeaves: true,
    wrapWithDirectory: true,
  }),
]);

export class IPNSFixture {
  public readonly ipnsId: string;

  constructor(ipnsId: string) {
    this.ipnsId = ipnsId;
  }

  public static async fromFixture(
    fixture: string,
    options: IPNSFixtureOptions
  ) {
    const ipnsId = await provisionIPNSWithKubo(fixture, options);
    return new IPNSFixture(ipnsId);
  }
}

const ipnsFixtures = await Promise.all([
  IPNSFixture.fromFixture(`/ipfs/${Fixture.get("root2").root.cid}`, {
    key: "cache_test_key",
    allowOffline: true,
  }),
]);

export const ipnsFixture = ipnsFixtures[0];
