import { Fixture } from "./util/fixtures.js";
import { provisioners } from "./util/provisioners.js";

async function provision(...args: string[]) {
  const [provisionerName, path, ..._rest] = args;

  const fixtures = Fixture.getAll().filter(
    (fixture) => path === undefined || path === fixture.path
  );

  const provisioner = provisioners[provisionerName];

  if (!provisioner) {
    throw new Error(`Unknown provisioner: ${provisionerName}`);
  }

  for (const fixture of fixtures) {
    await provisioner(fixture);
  }
}

provision(...process.argv.slice(2))
  .then(() => process.exit(0))
  .catch((err) => {
    console.error(err);
    process.exit(1);
  });
