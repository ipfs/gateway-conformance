import { Context } from './util/context.js';
import { transformFixtureToSource } from './util/transforms.js'

const IPLD_RAW_TYPE = 'application/vnd.ipld.raw';

const rawBlockTestSources = {
  'dir': {
    source: transformFixtureToSource('dir'),
    options: {
      cidVersion: 1,
      rawLeaves: true,
      wrapWithDirectory: true,
    }
  }
}

const rawBlockTestContext = await Context.fromSources(rawBlockTestSources)

export const rawBlockTest = {
  'Test HTTP Gateway Raw Block (application/vnd.ipld.raw) Support': {
    tests: {
      'GET with format=raw param returns a raw block': {
        url: `/ipfs/${rawBlockTestContext.get('dir').getRootCID()}/dir?format=raw`,
        expect: rawBlockTestContext.get('dir').getString('dir')
      },
      'GET for application/vnd.ipld.raw returns a raw block': {
        url: `/ipfs/${rawBlockTestContext.get('dir').getRootCID()}/dir`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: rawBlockTestContext.get('dir').getString('dir')
      },
      'GET response for application/vnd.ipld.raw has expected response headers': {
        url: `/ipfs/${rawBlockTestContext.get('dir').getRootCID()}/dir/ascii.txt`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: {
          headers: {
            'content-type': IPLD_RAW_TYPE,
            'content-length': rawBlockTestContext.get('dir').getLength('dir/ascii.txt').toString(),
            'content-disposition': new RegExp(`attachment;\\s*filename="${rawBlockTestContext.get('dir').getCID('dir/ascii.txt')}\\.bin`),
            'x-content-type-options': 'nosniff'
          },
          body: rawBlockTestContext.get('dir').getString('dir/ascii.txt')
        }
      },
      'GET for application/vnd.ipld.raw with query filename includes Content-Disposition with custom filename': {
        url: `/ipfs/${rawBlockTestContext.get('dir').getRootCID()}/dir/ascii.txt?filename=foobar.bin`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: {
          headers: {
            'content-disposition': new RegExp(`attachment;\\s*filename="foobar\\.bin"`)
          }
        }
      },
      'GET response for application/vnd.ipld.raw has expected caching headers': {
        url: `/ipfs/${rawBlockTestContext.get('dir').getRootCID()}/dir/ascii.txt`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: {
          headers: {
            'etag': `"${rawBlockTestContext.get('dir').getCID('dir/ascii.txt')}.raw"`,
            'x-ipfs-path': `/ipfs/${rawBlockTestContext.get('dir').getRootCID()}/dir/ascii.txt`,
            'x-ipfs-roots': new RegExp(rawBlockTestContext.get('dir').getRootCID()),
          }
        }
      }
    }
  }
}
