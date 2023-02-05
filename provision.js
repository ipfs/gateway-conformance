import { provisioners } from './util/provisioners.js'
import { suites } from './util/suites.js'

async function provision(provisioner, suite) {
  if (suite === undefined) {
    for (const suite of Object.keys(suites)) {
      const fixtures = suites[suite].fixtures
      for (const fixture of fixtures) {
        await provisioners[provisioner](fixture)
      }
    }
  } else {
    const fixtures = suites[suite].fixtures
    for (const fixture of fixtures) {
      await provisioners[provisioner](fixture)
    }
  }
}

provision(...process.argv.slice(2))
