import { GraphData as ForceGraphData } from "force-graph";
import { GraphModel, LinkModel, NodeModel } from "../models";

export class GraphData implements GraphModel {
  private readonly nodes = new Map<string, NodeModel>();
  private readonly links = new Set<LinkModel>();

  public data(): ForceGraphData {
    return {
      nodes: Array.from(this.nodes.values()),
      links: Array.from(this.links.values()),
    };
  }

  public removeNode(id: string): void {
    const node = this.nodes.get(id);
    if (!node) return;
    this.nodes.delete(id);
    console.debug("removed node", id);

    for (const link of this.links.values()) {
      if (link.source === node || link.target === node) {
        this.links.delete(link);
      }
    }
  }

  public getOrCreateNode(id: string): NodeModel {
    let node = this.nodes.get(id);
    if (!node) {
      node = { id } as NodeModel;
      this.nodes.set(id, node);
      console.debug("added node", id);
    }
    return node;
  }

  public getOrCreateLink(sourceID: string, targetID: string) {
    let link = this.findLink(sourceID, targetID);
    if (!link) {
      const source = this.getOrCreateNode(sourceID);
      const target = this.getOrCreateNode(targetID);
      link = { source, target };
      this.links.add(link);
      console.debug("added link", sourceID, '=>', targetID);
    }
    return link;
  }

  private findLink(sourceID: string, targetID: string): LinkModel | null {
    for (const link of this.links.values()) {
      if (link.source.id === sourceID && link.target.id === targetID) {
        return link;
      }
    }
    return null;
  }

  public listLinksFrom(sourceID: string): Set<LinkModel> {
    const links = new Set<LinkModel>();
    for (const link of this.links.values()) {
      if (link.source.id === sourceID) {
        links.add(link);
      }
    }
    return links;
  }

  public removeLink(link: LinkModel): void {
    if (!this.links.delete(link)) return;
    console.debug("removed link", link.source.id, "=>", link.target.id);
  }
}
