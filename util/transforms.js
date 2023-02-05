import * as fs from 'fs'
import { getAllFilesSync } from 'get-all-files'

export function transformFixtureToSource(path, options = {}) {
  const absolutePath = `${process.cwd()}/fixtures/${path}`
  if (fs.lstatSync(absolutePath).isDirectory()) {
    const source = []
    for (const file of getAllFilesSync(absolutePath)) {
      source.push({
        ...options,
        path: file.slice(`${process.cwd()}/fixtures`.length),
        content: fs.readFileSync(file)
      })
    }
    return source
  } else {
    return {
      ...options,
      path: `/${path}`,
      content: fs.readFileSync(absolutePath)
    }
  }
}
