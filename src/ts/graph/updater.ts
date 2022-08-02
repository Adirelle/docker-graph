import { GraphModel, LinkModel, NodeModel } from "../models";

export class Updater {
  private queue: NodeModel[] = [];
  private seen = new Set<NodeModel>();

  public constructor(
    private readonly graph: GraphModel
  ) { }

  public update(...nodes: NodeModel[]) {
    this.queue = [];
    this.seen.clear();
    for (const node of nodes) {
      this.consider(node);
    }
    this.processNodes();
  }

  private consider(node: NodeModel) {
    if (this.seen.has(node)) return;
    this.seen.add(node);
    this.queue.push(node);
  }

  private processNodes() {
    let current: NodeModel | undefined;
    while (current = this.queue.shift()) {
      let node = this.graph.getOrCreateNode(current.id);
      Object.assign(node, current);
      delete node.links;

      this.processLinks(node, current.links || []);
    }
  }

  private processLinks(source: NodeModel, targets: NodeModel[]): void {
    const toRemove = new Set<LinkModel>(this.graph.listLinksFrom(source.id));

    for (const target of targets) {
      const link = this.graph.getOrCreateLink(source.id, target.id);
      toRemove.delete(link);
      this.consider(target);
    }

    toRemove.forEach(l => this.graph.removeLink(l));
  }
}
