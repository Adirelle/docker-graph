import { LinkObject, NodeObject } from "force-graph";

export interface Hideable {
  isVisible(): boolean;
}

export interface Renderable extends Hideable {
  render(ctx: CanvasRenderingContext2D, scale: number): void;
  paintInteractionArea(
    color: string,
    ctx: CanvasRenderingContext2D,
    scale: number
  ): void;
}

export interface Link<A, B> extends Hideable, LinkObject {
  readonly source: Node<A>;
  readonly target: Node<B>;

}

export type LinkFrom<T> = Link<T, any>;
export type LinkTo<T> = Link<any, T>;

export interface Node<T> extends Renderable, NodeObject {
  readonly id: ID<T>;

  updateFrom(data: T, ctx: Context): boolean;
}

export type AnyNode = Node<any>;
export type AnyLink = Link<any, any>;

export type ID<T> = string & { readonly brand?: T; };

export interface Context {
  getOrCreateNode<N extends Node<T>, T>(
    id: ID<T>,
    cons: (data: T) => N,
    data: T
  ): N;

  removeNode<T>(node: Node<T>): boolean;

  getOrCreateLink<A, B>(source: Node<A>, dest: Node<B>): Link<A, B>;
  removeLink(link: AnyLink): boolean;

  listLinksFrom<T>(source: Node<T>): Iterable<LinkFrom<T>>;
}
