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

  static async fromPath(path, options) {
    const absolute = new URL(joinPath('..', 'fixtures', path), import.meta.url).pathname
    const source = []
    const isDirectory = fs.lstatSync(absolute).isDirectory()
    if (isDirectory) {
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
