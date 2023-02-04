import fs from 'fs'
import { importer } from 'ipfs-unixfs-importer'
import { exporter } from 'ipfs-unixfs-exporter'
import { getAllFilesSync } from 'get-all-files'
import { MemoryBlockstore } from 'blockstore-core/memory'
import * as dagPB from '@ipld/dag-pb'

const IPLD_RAW_TYPE = 'application/vnd.ipld.raw';

const blockstore = new MemoryBlockstore()
const source = []
for (const file of getAllFilesSync(`${process.cwd()}/fixtures/dir`)) {
  source.push({
    path: file.slice(`${process.cwd()}/fixtures`.length),
    content: fs.createReadStream(file)
  })
}
const importerOptions = {
  cidVersion: 1,
  rawLeaves: true,
  wrapWithDirectory: true,
}
const exporterOptions = {}
const files = []
for await (const file of importer(source, blockstore, importerOptions)) {
  const entry = await exporter(file.cid, blockstore, exporterOptions)
  let buffer
  if (entry.type === 'raw') {
    buffer = entry.node
  } else {
    buffer = Buffer.from(dagPB.encode(entry.node))
  }
  files.push({
    file, entry, buffer
  })
}

function getFile(path) {
  return files.find(f => f.file.path === path)
}

export const rawBlockTest = {
  'Test HTTP Gateway Raw Block (application/vnd.ipld.raw) Support': {
    tests: {
      'GET with format=raw param returns a raw block': {
        url: `/ipfs/${getFile('').file.cid}/dir?format=raw`,
        expect: getFile('dir').buffer.toString()
      },
      'GET for application/vnd.ipld.raw returns a raw block': {
        url: `/ipfs/${getFile('').file.cid}/dir`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: getFile('dir').buffer.toString()
      },
      'GET response for application/vnd.ipld.raw has expected response headers': {
        url: `/ipfs/${getFile('').file.cid}/dir/ascii.txt`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: {
          headers: {
            'content-type': IPLD_RAW_TYPE,
            'content-length': getFile('dir/ascii.txt').buffer.length.toString(),
            'content-disposition': new RegExp(`attachment;\\s*filename="${getFile('dir/ascii.txt').file.cid}\\.bin`),
            'x-content-type-options': 'nosniff'
          },
          body: getFile('dir/ascii.txt').buffer.toString()
        }
      },
      'GET for application/vnd.ipld.raw with query filename includes Content-Disposition with custom filename': {
        url: `/ipfs/${getFile('').file.cid}/dir/ascii.txt?filename=foobar.bin`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: {
          headers: {
            'content-disposition': new RegExp(`attachment;\\s*filename="foobar\\.bin"`)
          }
        }
      },
      'GET response for application/vnd.ipld.raw has expected caching headers': {
        url: `/ipfs/${getFile('').file.cid}/dir/ascii.txt`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: {
          headers: {
            'etag': `"${getFile('dir/ascii.txt').file.cid}.raw"`,
            'x-ipfs-path': `/ipfs/${getFile('').file.cid}/dir/ascii.txt`,
            'x-ipfs-roots': new RegExp(getFile('').file.cid),
          }
        }
      }
    }
  }
}
