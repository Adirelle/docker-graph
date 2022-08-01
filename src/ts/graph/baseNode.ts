import { Context, Node } from "./types";

interface DataType {
  ID: string;
}

export abstract class BaseNode<T extends DataType> implements Node<T> {
  public id: string;
  public x?: number;
  public y?: number;
  public vx?: number;
  public vy?: number;
  public fx?: number;
  public fy?: number;
  public label?: string;

  public constructor(data: T) {
    this.id = data.ID;
  }

  public updateFrom(_data: T, _ctx: Context): boolean {
    return false;
  }

  public isVisible(): boolean {
    return true;
  }

  public abstract render(ctx: CanvasRenderingContext2D, scale: number): void;

  public paintInteractionArea(
    color: string,
    ctx: CanvasRenderingContext2D,
    _: number
  ): void {
    const { x, y } = this;
    if (!x || !y) return;

    ctx.fillStyle = color;
    ctx.fillRect(x - 5, y - 5, 10, 10);
  }
}
