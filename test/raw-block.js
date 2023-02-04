import { rawBlockTest } from "../raw-block.js";
import { api, run } from 'declarative-e2e-test';

// https://marc-ed-raffalli.github.io/declarative-e2e-test
run(rawBlockTest, {
  api: api.mocha,
  config: {
    url: process.env.GATEWAY_URL || 'http://localhost:8080'
  },
  logLevel: process.env.LOG_LEVEL || 'SILENT'
})
