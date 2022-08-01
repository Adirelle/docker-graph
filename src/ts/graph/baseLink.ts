import { Link, Node } from "./types";

export class BaseLink<A, B> implements Link<A, B> {
  public constructor(
    public readonly source: Node<A>,
    public readonly target: Node<B>
  ) {
    this.source = source;
    this.target = target;
  }

  public isVisible(): boolean {
    return this.source.isVisible() && this.target.isVisible();
  }
}
