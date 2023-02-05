import {Collection} from './collection.js'

export class Context {
  constructor(collections) {
    this.collections = collections
  }

  static async fromFixtures(fixtures) {
    const collections = {}
    for (const [name, {paths, options}] of Object.entries(fixtures)) {
      collections[name] = await Collection.fromPaths(paths, options)
    }
    return new Context(collections)
  }

  getFixture(name) {
    return this.collections[name]
  }
}
