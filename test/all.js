import { suites } from '../util/suites.js'
import { api, run } from 'declarative-e2e-test';

const config = {
  api: api.mocha,
  config: {
    url: process.env.GATEWAY_URL || 'http://localhost:8080'
  },
  logLevel: process.env.LOG_LEVEL || 'SILENT'
}

// https://marc-ed-raffalli.github.io/declarative-e2e-test
for (const [_name, { _fixtures, test }] of Object.entries(suites)) {
  run(test, config)
}
