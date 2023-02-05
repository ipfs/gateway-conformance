import {Collection} from './collection.js'

export class Context {
  constructor(collections) {
    this.collections = collections
  }

  static async fromSources(sources) {
    const collections = {}
    for (const [name, {source, options}] of Object.entries(sources)) {
      collections[name] = await Collection.fromSource(source, options)
    }
    return new Context(collections)
  }

  get(name) {
    return this.collections[name]
  }
}
