import { NodeModel } from "../graph/nodes/baseNode";
import { AnyNode, Context, Node } from "../graph/types";

export abstract class LinkCollection<I, M extends NodeModel, N extends Node<M>> {

  public constructor(public readonly source: AnyNode) { }

  public update(items: Iterable<I>, ctx: Context): boolean {
    let changed = false;

    const nodes = new Set<N>;

    for (const item of items) {
      if (!this.accept(item)) continue;
      const model = this.mapModel(item);
      const node = ctx.getOrCreateNode(model.ID, (m) => this.buildNode(m), model);
      ctx.getOrCreateLink(this.source, node);
      nodes.add(node);
    }

    for (const link of ctx.listLinksFrom(this.source)) {
      if (this.isTargetNode(link.target) && !nodes.has(link.target)) {
        ctx.removeLink(link);
        changed = true;
      }
    }

    return changed;
  }

  protected accept(_item: I): boolean {
    return true;
  }

  protected abstract mapModel(item: I): M;

  protected abstract buildNode(model: M): N;

  protected abstract isTargetNode(node: AnyNode): node is N;
}

