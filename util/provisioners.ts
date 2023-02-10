import axios from "axios";
import { execSync } from "child_process";
import { Fixture } from "./fixtures";

type ProvisionFct = (fixture: Fixture) => Promise<void>;

export const provisioners: { [key: string]: ProvisionFct } = {
  kubo: provisionWithKubo,
  "writeable-gateway": provisionWithWriteableGateway,
};

export interface IPNSFixtureOptions {
  key: string;
  allowOffline: boolean;
}

export async function provisionIPNSWithKubo(
  path: string,
  options: IPNSFixtureOptions
) {
  const args: string[] = [];

  // generate the key
  const outGen = execSync(
    `ipfs key gen --ipns-base=base36 --type=ed25519 ${options.key} | head -n1 | tr -d "\n"`
  );
  const ipnsId = outGen.toString();
  console.log(`ipnsId: ${ipnsId}`);

  // publish the ipns record
  args.push("--key", options.key);

  if (options.allowOffline) {
    args.push("--allow-offline");
  }

  console.log(`RUNNING: ipfs name publish ${args.join(" ")} "${path}"`);
  const out = execSync(`ipfs name publish ${args.join(" ")} "${path}"`);
  console.log(out.toString());

  // test the ipns record
  console.log(`RUNNING: ipfs name resolve "${ipnsId}"`)
  const outTest = execSync(`ipfs name resolve "${ipnsId}"`);
  console.log(`out: ${outTest.toString()}`);

  return ipnsId;
}

async function provisionWithKubo(fixture: Fixture) {
  const args: string[] = [];
  if (fixture.isDirectory()) {
    args.push("--recursive");
  }
  if (fixture.options.rawLeaves) {
    args.push("--raw-leaves");
  }
  if (fixture.options.cidVersion) {
    args.push("--cid-version", fixture.options.cidVersion.toString());
  }
  if (fixture.options.wrapWithDirectory) {
    args.push("--wrap-with-directory");
  }

  const out = execSync(
    `ipfs add ${args.join(" ")} ${fixture.getAbsolutePath()}`
  );

  console.log(out.toString());
}

async function provisionWithWriteableGateway(fixture: Fixture) {
  const baseURL = process.env.GATEWAY_URL || "http://localhost:8080";
  const client = axios.create({ baseURL });

  const response = await client.post("/ipfs/", fixture.root.raw, {
    headers: { "Content-Type": "application/vnd.ipld.raw" },
    maxRedirects: 0,
  });

  if (response.status !== 201) {
    throw new Error(`Unexpected status code: ${response.status}`);
  }
  if (response.headers["ipfs-hash"] !== fixture.cid.toString()) {
    throw new Error(
      `Unexpected IPFS hash: ${response.headers["ipfs-hash"]} (expected ${fixture.cid})`
    );
  }
}
