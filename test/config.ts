import { api, ITestConfig } from "declarative-e2e-test";

const isLogLevel = (
  level: string | undefined
): level is ITestConfig["logLevel"] => {
  return ["SILENT", "ERROR", "DEBUG", "TRACE", undefined].includes(level);
};

const assertLogLevel = (level: string | undefined): ITestConfig["logLevel"] => {
  if (!isLogLevel(level)) {
    throw new Error(`Invalid log level: ${level}`);
  }

  return level;
};

export const config: ITestConfig = {
  api: api.mocha,
  config: {
    url: process.env.GATEWAY_URL || "http://localhost:8080",
  },
  logLevel: assertLogLevel(process.env.LOG_LEVEL || "SILENT"),
};
