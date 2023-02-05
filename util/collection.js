import * as fs from 'fs'
import { getAllFilesSync } from 'get-all-files'
import { importer } from 'ipfs-unixfs-importer'
import { exporter } from 'ipfs-unixfs-exporter'
import { MemoryBlockstore } from 'blockstore-core/memory'
import * as dagPB from '@ipld/dag-pb'

const ROOT_PATH = `${process.cwd()}/fixtures`

export class Collection {
  constructor(entries) {
    this.entries = entries
  }

  static async fromPaths(paths, options = {}) {
    const allPaths = []
    for (const path of paths) {
      if (fs.lstatSync(`${ROOT_PATH}/${path}`).isDirectory()) {
        allPaths.push(...getAllFilesSync(`${ROOT_PATH}/${path}`))
      } else {
        allPaths.push(`${ROOT_PATH}/${path}`)
      }
    }
    const source = []
    for (const path of allPaths) {
      source.push({
        path: path.slice(`${ROOT_PATH}/`.length),
        content: fs.readFileSync(path)
      })
    }
    const blockstore = new MemoryBlockstore()
    const entries = []
    for await (const importResult of importer(source, blockstore, options)) {
      const exportResult = await exporter(importResult.cid, blockstore)
      let raw
      if (exportResult.type === 'raw') {
        raw = exportResult.node
      } else {
        raw = Buffer.from(dagPB.encode(exportResult.node))
      }
      entries.push({
        importResult,
        exportResult,
        raw
      })
    }
    return new Collection(entries)
  }

  get(path) {
    return this.entries.find(entry => entry.importResult.path === path)
  }

  getCID(path) {
    return this.get(path).importResult.cid
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
