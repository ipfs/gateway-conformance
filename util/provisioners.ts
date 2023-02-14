import axios from "axios";
import { execSync } from "child_process";
import fs from "fs";

type ProvisionFct = () => Promise<void>;

export const provisioners: { [key: string]: ProvisionFct } = {
  kubo: provisionWithKubo,
  "writeable-gateway": provisionWithWriteableGateway,
};

async function provisionWithKubo() {
  const out = execSync("ipfs dag import ./fixtures.car");
  console.log(out.toString());
}

async function provisionWithWriteableGateway() {
  const baseURL = process.env.GATEWAY_URL || "http://127.0.0.1:8080";
  const client = axios.create({ baseURL });

  const data = fs.readFileSync("./fixtures.car");

  const response = await client.post("/ipfs/", data, {
    headers: { "Content-Type": "application/vnd.ipfs.car" },
    maxRedirects: 0,
  });

  if (response.status !== 201) {
    throw new Error(`Unexpected status code: ${response.status}`);
  }

  console.log(response.data.toString());
}
