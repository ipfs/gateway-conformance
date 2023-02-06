import * as fs from 'fs'
import { join as joinPath } from 'path'
import { getAllFilesSync } from 'get-all-files'
import { importer } from 'ipfs-unixfs-importer'
import { exporter } from 'ipfs-unixfs-exporter'
import { MemoryBlockstore } from 'blockstore-core/memory'
import * as dagPB from '@ipld/dag-pb'

export class Fixture {
  constructor(path, options, entries) {
    this.path = path;
    this.options = options;
    this.entries = entries;
  }

  static get(path) {
    return fixtures.find(fixture => fixture.path === path)
  }

  static getAll() {
    return fixtures
  }

  static getAbsolutePath(path) {
    return new URL(joinPath('..', 'fixtures', path), import.meta.url).pathname
  }

  getAbsolutePath() {
    return Fixture.getAbsolutePath(this.path)
  }

  static isDirectory(path) {
    return fs.lstatSync(Fixture.getAbsolutePath(path)).isDirectory()
  }

  isDirectory() {
    return Fixture.isDirectory(this.path)
  }

  static async fromPath(path, options) {
    const absolute = Fixture.getAbsolutePath(path)
    const source = []
    if (Fixture.isDirectory(path)) {
      for (const file of getAllFilesSync(absolute)) {
        source.push({
          path: `${path}/${file.slice(`${absolute}/`.length)}`,
          content: fs.readFileSync(file)
        })
      }
    } else {
      source.push({
        path,
        content: fs.readFileSync(absolute)
      })
    }
    const blockstore = new MemoryBlockstore()
    const entries = []
    for await (const imported of importer(source, blockstore, options)) {
      const exported = await exporter(imported.cid, blockstore)
      let raw
      if (exported.type === 'raw') {
        raw = exported.node
      } else {
        raw = Buffer.from(dagPB.encode(exported.node))
      }
      entries.push({
        imported,
        exported,
        raw
      })
    }
    return new Fixture(path, options, entries)
  }

  get(path) {
    return this.entries.find(entry => entry.imported.path === path)
  }

  getRoot() {
    return this.get('')
  }

  getCID(path) {
    return this.get(path).imported.cid
  }

  getRootCID() {
    return this.getCID('')
  }

  getRaw(path) {
    return this.get(path).raw
  }

  getString(path) {
    return this.getRaw(path).toString()
  }

  getLength(path) {
    return this.getRaw(path).length
  }
}

const fixtures = [
  await Fixture.fromPath('dir', {
    cidVersion: 1,
    rawLeaves: true,
    wrapWithDirectory: true,
  })
]
