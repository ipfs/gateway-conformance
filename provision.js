import { Fixture } from './util/fixtures.js'
import { provisioners } from './util/provisioners.js'

async function provision(provisioner, path) {
  const fixtures = Fixture.getAll().filter(fixture => path === undefined || path === fixture.path)
  for (const fixture of fixtures) {
    await provisioners[provisioner](fixture)
  }
}

provision(...process.argv.slice(2))
