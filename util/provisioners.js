import { execSync } from 'child_process'
import axios from 'axios'

export const provisioners = {
  'kubo': provisionWithKubo,
  'writeable-gateway': provisionWithWriteableGateway,
}

async function provisionWithKubo(fixture) {
  const args = []
  if (fixture.isDirectory()) {
    args.push('--recursive')
  }
  if (fixture.options.rawLeaves) {
    args.push('--raw-leaves')
  }
  if (fixture.options.cidVersion) {
    args.push('--cid-version', fixture.options.cidVersion)
  }
  if (fixture.options.wrapWithDirectory) {
    args.push('--wrap-with-directory')
  }
  const out = execSync(`ipfs add ${args.join(' ')} ${fixture.getAbsolutePath()}`)
  console.log(out.toString())
}

async function provisionWithWriteableGateway(fixture) {
  const baseURL = process.env.GATEWAY_URL || 'http://localhost:8080'
  const client = axios.create({ baseURL })
  const response = await client.post('/ipfs/', fixture.getRoot().raw, {
    headers: { 'Content-Type': 'application/vnd.ipld.raw' },
    maxRedirects: 0,
  });
  if (response.status !== 201) {
    throw new Error(`Unexpected status code: ${response.status}`)
  }
  if (response.headers['ipfs-hash'] !== fixture.getRootCID().toString()) {
    throw new Error(`Unexpected IPFS hash: ${response.headers['ipfs-hash']} (expected ${fixture.getRootCID()})`)
  }
}
