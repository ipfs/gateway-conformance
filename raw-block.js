import { Fixture } from './util/fixture.js';

const IPLD_RAW_TYPE = 'application/vnd.ipld.raw';

export const rawBlockTestFixtures = [
  await Fixture.fromPath('dir', {
    cidVersion: 1,
    rawLeaves: true,
    wrapWithDirectory: true,
  })
]

function getFixture(path) {
  return rawBlockTestFixtures.find(fixture => fixture.path === path)
}

export const rawBlockTest = {
  'Test HTTP Gateway Raw Block (application/vnd.ipld.raw) Support': {
    tests: {
      'GET with format=raw param returns a raw block': {
        url: `/ipfs/${getFixture('dir').getRootCID()}/dir?format=raw`,
        expect: [200, getFixture('dir').getString('dir')]
      },
      'GET for application/vnd.ipld.raw returns a raw block': {
        url: `/ipfs/${getFixture('dir').getRootCID()}/dir`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: [200, getFixture('dir').getString('dir')]
      },
      'GET response for application/vnd.ipld.raw has expected response headers': {
        url: `/ipfs/${getFixture('dir').getRootCID()}/dir/ascii.txt`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: [200, {
          headers: {
            'content-type': IPLD_RAW_TYPE,
            'content-length': getFixture('dir').getLength('dir/ascii.txt').toString(),
            'content-disposition': new RegExp(`attachment;\\s*filename="${getFixture('dir').getCID('dir/ascii.txt')}\\.bin`),
            'x-content-type-options': 'nosniff'
          },
          body: getFixture('dir').getString('dir/ascii.txt')
        }]
      },
      'GET for application/vnd.ipld.raw with query filename includes Content-Disposition with custom filename': {
        url: `/ipfs/${getFixture('dir').getRootCID()}/dir/ascii.txt?filename=foobar.bin`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: [200, {
          headers: {
            'content-disposition': new RegExp(`attachment;\\s*filename="foobar\\.bin"`)
          }
        }]
      },
      'GET response for application/vnd.ipld.raw has expected caching headers': {
        url: `/ipfs/${getFixture('dir').getRootCID()}/dir/ascii.txt`,
        headers: {accept: IPLD_RAW_TYPE},
        expect: [200, {
          headers: {
            'etag': `"${getFixture('dir').getCID('dir/ascii.txt')}.raw"`,
            'x-ipfs-path': `/ipfs/${getFixture('dir').getRootCID()}/dir/ascii.txt`,
            'x-ipfs-roots': new RegExp(getFixture('dir').getRootCID())
          }
        }]
      }
    }
  }
}
