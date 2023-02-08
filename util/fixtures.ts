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

interface FixtureOptions extends UserImporterOptions {}

interface Entry {
  imported: ImportResult;
  exported: UnixFSEntry;
  raw: Buffer | Uint8Array;
}

export class Fixture {
  public readonly path: string;
  public readonly options: FixtureOptions;
  public readonly entries: Entry[];

  constructor(path: string, options: FixtureOptions, entries: Entry[]) {
    this.path = path;
    this.options = options;
    this.entries = entries;
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
      entries.push({
        imported,
        exported,
        raw,
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

  getRoot(): Entry {
    return this.get("");
  }

  getCID(path: string): CID {
    return this.get(path).imported?.cid;
  }

  getRootCID(): CID {
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

const fixtures = await Promise.all([
  Fixture.fromPath("dir", {
    cidVersion: 1,
    rawLeaves: true,
    wrapWithDirectory: true,
  }),
]);
