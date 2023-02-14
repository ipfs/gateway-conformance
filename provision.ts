import { provisioners } from "./util/provisioners.js";

async function provision(...args: string[]) {
  const [provisionerName, path, ..._rest] = args;

  const provisioner = provisioners[provisionerName];

  if (!provisioner) {
    throw new Error(`Unknown provisioner: ${provisionerName}`);
  }

  await provisioner();
}

provision(...process.argv.slice(2))
  .then(() => process.exit(0))
  .catch((err) => {
    console.error(err);
    process.exit(1);
  });
