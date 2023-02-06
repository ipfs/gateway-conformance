import { api } from 'declarative-e2e-test';

export const config = {
  api: api.mocha,
  config: {
    url: process.env.GATEWAY_URL || 'http://localhost:8080'
  },
  logLevel: process.env.LOG_LEVEL || 'SILENT'
}
