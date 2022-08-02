import { NodeUpdateFunc, Updater } from "../eventProcessor";
import { GraphModel, LinkModel, NodeModel } from "../models";

export class GraphUpdater implements Updater {
  private readonly visitedNodes = new Map<string, NodeModel>();
  private readonly unvisitedLinks = new Map<string, Set<LinkModel>>();

  public constructor(
    private readonly graph: GraphModel
  ) { }

  public updateNode(id: string, update: NodeUpdateFunc): void {
    if (this.visitedNodes.has(id)) return;
    const node = this.graph.getOrCreateNode(id);
    this.visitedNodes.set(id, node);
    update(node, this);
  }

  public removeNode(id: string): void {
    this.graph.removeNode(id);
  }

  public updateLink(sourceID: string, targetID: string, update: NodeUpdateFunc): void {
    let unvisitedLinks = this.unvisitedLinks.get(sourceID);
    if (!unvisitedLinks) {
      unvisitedLinks = this.graph.listLinksFrom(sourceID);
      this.unvisitedLinks.set(sourceID, unvisitedLinks);
    }
    const link = this.graph.getOrCreateLink(sourceID, targetID);
    unvisitedLinks.delete(link);
    this.updateNode(targetID, update);
  }

  public tidy() {
    for (const unvisitedLinks of this.unvisitedLinks.values()) {
      for (const link of unvisitedLinks.values()) {
        this.graph.removeLink(link);
      }
    }
  }
}
