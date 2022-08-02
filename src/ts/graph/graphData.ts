import { GraphData as ForceGraphData } from "force-graph";
import { Event } from "../api";
import { BaseLink } from "./baseLink";
import { ContainerNode } from "./nodes/container";
import { AnyLink, AnyNode, Context, ID, Link, LinkFrom, Node } from "./types";

export class GraphData implements Context {
  private readonly nodes = new Map<ID<any>, AnyNode>();
  private readonly links = new Set<AnyLink>();

  private dirty = false;

  public constructor() { }

  public data(): ForceGraphData {
    this.dirty = false;
    return {
      nodes: Array.from(this.nodes.values()),
      links: Array.from(this.links.values()),
    };
  }

  public process(event: Event): boolean {
    if (event.TargetType != "container") {
      return false;
    }
    if (event.Type == "removed") {
      this.removeNodeByID(event.TargetID);
    } else {
      this.getOrCreateNode(event.TargetID, (m) => new ContainerNode(m), event.Details);
    }
    return this.dirty;
  }

  public markDirty(): void {
    if (!this.dirty) {
      console.debug("graph is dirty");
      this.dirty = true;
    }
  }

  public isDirty(): boolean {
    return this.dirty;
  }

  public removeNode<T>(node: Node<T>): boolean {
    return this.removeNodeByID(node.id);
  }

  private removeNodeByID<T>(id: ID<T>): boolean {
    const node = this.nodes.get(id);
    if (!node) return false;
    this.nodes.delete(id);
    console.debug("removed node", node);
    this.markDirty();

    for (const link of this.links.values()) {
      if (link.source === node || link.target === node) {
        this.links.delete(link);
      }
    }
    return true;
  }

  public getOrCreateNode<N extends Node<T>, T>(
    id: ID<T>,
    builder: (data: T) => N,
    data: T
  ): N {
    let node = this.nodes.get(id) as N;
    if (!node) {
      node = builder(data);
      this.nodes.set(id, node);
      console.debug("new node", node);
      this.markDirty();
    }
    if (node.updateFrom(data, this)) {
      this.markDirty();
    }
    return node;
  }

  public getOrCreateLink<A, B>(source: Node<A>, target: Node<B>): Link<A, B> {
    const existing = this.findLink(source, target);
    if (existing) {
      return existing;
    }
    const link = new BaseLink(source, target);
    this.links.add(link);
    console.debug("new link", link);
    this.markDirty();
    return link;
  }

  private findLink<A, B>(source: Node<A>, target: Node<B>): Link<A, B> | null {
    for (const link of this.links.values()) {
      if (link.source === source && link.target === target) {
        return link;
      }
    }
    return null;
  }

  public *listLinksFrom<T>(source: Node<T>): Iterable<LinkFrom<T>> {
    for (const link of this.links.values()) {
      if (link.source === source) {
        yield link;
      }
    }
  }

  public removeLink(link: AnyLink): boolean {
    if (!this.links.delete(link)) return false;

    console.debug("removed link", link);
    this.markDirty();
    return true;
  }
}
