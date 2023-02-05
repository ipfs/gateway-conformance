import { Context } from './util/context.js';

const IPLD_RAW_TYPE = 'application/vnd.ipld.raw';

const rawBlockTestFixtures = {
  'dir': {
    paths: ['dir'],
    options: {
      cidVersion: 1,
      rawLeaves: true,
      wrapWithDirectory: true,
    }
  }
}

export const rawBlockTest = {
  'Test HTTP Gateway Raw Block (application/vnd.ipld.raw) Support': {
    before: async function() {
      this.context = await Context.fromFixtures(rawBlockTestFixtures)
    },
    tests: {
      'GET with format=raw param returns a raw block': {
        url: function() {
          return `/ipfs/${this.context.getFixture('dir').getRootCID()}/dir?format=raw`
        },
        expect: [200, function() {
          return this.context.getFixture('dir').getString('dir')
        }]
      },
      'GET for application/vnd.ipld.raw returns a raw block': {
        url: function() {
          return `/ipfs/${this.context.getFixture('dir').getRootCID()}/dir`
        },
        headers: {accept: IPLD_RAW_TYPE},
        expect: [200, function() {
          return this.context.getFixture('dir').getString('dir')
        }]
      },
      'GET response for application/vnd.ipld.raw has expected response headers': {
        url: function() {
          return `/ipfs/${this.context.getFixture('dir').getRootCID()}/dir/ascii.txt`
        },
        headers: {accept: IPLD_RAW_TYPE},
        expect: [200, function() {
          return {
            headers: {
              'content-type': IPLD_RAW_TYPE,
              'content-length': this.context.getFixture('dir').getLength('dir/ascii.txt').toString(),
              'content-disposition': new RegExp(`attachment;\\s*filename="${this.context.getFixture('dir').getCID('dir/ascii.txt')}\\.bin`),
              'x-content-type-options': 'nosniff'
            },
            body: this.context.getFixture('dir').getString('dir/ascii.txt')
          }
        }]
      },
      'GET for application/vnd.ipld.raw with query filename includes Content-Disposition with custom filename': {
        url: function() {
          return `/ipfs/${this.context.getFixture('dir').getRootCID()}/dir/ascii.txt?filename=foobar.bin`
        },
        headers: {accept: IPLD_RAW_TYPE},
        expect: [200, {
          headers: {
            'content-disposition': new RegExp(`attachment;\\s*filename="foobar\\.bin"`)
          }
        }]
      },
      'GET response for application/vnd.ipld.raw has expected caching headers': {
        url: function() {
          return `/ipfs/${this.context.getFixture('dir').getRootCID()}/dir/ascii.txt`
        },
        headers: {accept: IPLD_RAW_TYPE},
        expect: [200, function() {
          return {
            headers: {
              'etag': `"${this.context.getFixture('dir').getCID('dir/ascii.txt')}.raw"`,
              'x-ipfs-path': `/ipfs/${this.context.getFixture('dir').getRootCID()}/dir/ascii.txt`,
              'x-ipfs-roots': new RegExp(this.context.getFixture('dir').getRootCID())
            }
          }
        }]
      }
    }
  }
}
