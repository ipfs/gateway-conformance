import axios from "axios";
import { execSync } from "child_process";
import { Fixture } from "./fixtures";

type ProvisionFct = (fixture: Fixture) => Promise<void>;

export const provisioners: { [key: string]: ProvisionFct } = {
  kubo: provisionWithKubo,
  "writeable-gateway": provisionWithWriteableGateway,
};

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

  const response = await client.post("/ipfs/", fixture.getRoot().raw, {
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
